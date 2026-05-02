package core

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ksroido/athena/internal/db"
	"github.com/ksroido/athena/internal/hr"
)

// AgentStatus represents the current status of an Agent
type AgentStatus string

const (
	StatusIdle      AgentStatus = "idle"
	StatusWorking   AgentStatus = "working"
	StatusInMeeting AgentStatus = "in_meeting"
	StatusOffline   AgentStatus = "offline"
)

// AgentHandle wraps a running Agent goroutine
type AgentHandle struct {
	ID           string
	Role         string
	ProjectID    string
	Status       AgentStatus
	TaskCh       chan *TaskMessage
	SteerCh      chan string
	RestartCount int
	LastActive   time.Time
	cancel       context.CancelFunc
}

// TaskMessage is a message sent to an agent
type TaskMessage struct {
	TaskID    string
	Content   string
	FromAgent string
}

// AgentManager manages all Agent goroutines (replaces subprocess Supervisor)
type AgentManager struct {
	agents    map[string]*AgentHandle
	mu        sync.RWMutex
	llmConfig *LLMClient
	mainDB    *db.DB
	dataDir   string
	hr        *hr.HR
	hireFunc  func(req *hr.HireRequest) (*db.Agent, error)
}

// NewAgentManager creates a new AgentManager
func NewAgentManager(llm *LLMClient, mainDB *db.DB, dataDir string) *AgentManager {
	am := &AgentManager{
		agents:    make(map[string]*AgentHandle),
		llmConfig: llm,
		mainDB:    mainDB,
		dataDir:   dataDir,
	}
	return am
}

// SetHR sets the HR instance and initializes the hire callback
func (am *AgentManager) SetHR(h *hr.HR) {
	am.hr = h
	am.hireFunc = h.Hire
}

// StartAgent launches a new Agent goroutine
func (am *AgentManager) StartAgent(agent *db.Agent, projectID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	if _, exists := am.agents[agent.ID]; exists {
		return fmt.Errorf("agent %s is already running", agent.ID)
	}

	ctx, cancel := context.WithCancel(context.Background())

	handle := &AgentHandle{
		ID:         agent.ID,
		Role:       agent.Role,
		ProjectID:  projectID,
		Status:     StatusIdle,
		TaskCh:     make(chan *TaskMessage, 16),
		SteerCh:    make(chan string, 16),
		LastActive: time.Now(),
		cancel:     cancel,
	}

	// Create agent loop config — inject callbacks and DB
	loopCfg := &AgentLoopConfig{
		AgentID:      agent.ID,
		Role:         agent.Role,
		ProjectID:    projectID,
		DataDir:      am.dataDir,
		LLM:          am.llmConfig,
		MainDB:       am.mainDB,
		TaskFunc:     am.SendTask,
		HireFunc:     am.hireFunc,
		NotifyPMFunc: am.NotifyPM,
	}

	loop := NewAgentLoop(loopCfg)

	// Start agent goroutine
	go func() {
		if err := loop.RunInProcess(ctx, handle); err != nil {
			log.Printf("[agent:%s] exited: %v", agent.ID, err)
		}
		am.mu.Lock()
		if _, exists := am.agents[agent.ID]; exists {
			handle.Status = StatusOffline
		}
		am.mu.Unlock()
	}()

	am.agents[agent.ID] = handle
	log.Printf("[agent-manager] started %s (role=%s, project=%s)", agent.ID, agent.Role, projectID)

	return nil
}

// StartAgentFromHR implements hr.AgentStarter interface
func (am *AgentManager) StartAgentFromHR(agent *db.Agent, projectID string) error {
	return am.StartAgent(agent, projectID)
}

// StopAgent terminates an Agent goroutine
func (am *AgentManager) StopAgent(agentID string) error {
	am.mu.Lock()
	defer am.mu.Unlock()

	handle, exists := am.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	handle.cancel()
	delete(am.agents, agentID)
	return nil
}

// SendTask sends a task to an agent
func (am *AgentManager) SendTask(agentID, taskID, content, fromAgent string) error {
	am.mu.RLock()
	handle, exists := am.agents[agentID]
	am.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	handle.Status = StatusWorking
	handle.LastActive = time.Now()
	handle.TaskCh <- &TaskMessage{
		TaskID:    taskID,
		Content:   content,
		FromAgent: fromAgent,
	}

	return nil
}

// SendSteer sends a steering message to an agent
func (am *AgentManager) SendSteer(agentID, content string) error {
	am.mu.RLock()
	handle, exists := am.agents[agentID]
	am.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	handle.SteerCh <- content
	return nil
}

// GetAgent returns a running agent handle
func (am *AgentManager) GetAgent(agentID string) (*AgentHandle, bool) {
	am.mu.RLock()
	defer am.mu.RUnlock()

	handle, exists := am.agents[agentID]
	return handle, exists
}

// ListAgents returns all running agent handles
func (am *AgentManager) ListAgents() []*AgentHandle {
	am.mu.RLock()
	defer am.mu.RUnlock()

	result := make([]*AgentHandle, 0, len(am.agents))
	for _, handle := range am.agents {
		result = append(result, handle)
	}
	return result
}

// DataDir returns the data directory
func (am *AgentManager) DataDir() string {
	return am.dataDir
}

// MainDB returns the main database
func (am *AgentManager) MainDB() *db.DB {
	return am.mainDB
}

// LLMConfigValue returns the LLM client
func (am *AgentManager) LLMConfigValue() *LLMClient {
	return am.llmConfig
}

// NotifyPM finds the PM agent for a project and sends it a steer message.
// This is used by the submit_for_review tool to wake up PM for verification.
func (am *AgentManager) NotifyPM(projectID, message string) error {
	am.mu.RLock()
	defer am.mu.RUnlock()

	for _, handle := range am.agents {
		if handle.ProjectID == projectID && handle.Role == "pm" && handle.Status != StatusOffline {
			handle.SteerCh <- message
			log.Printf("[agent-manager] notified PM %s for project %s", handle.ID, projectID)
			return nil
		}
	}

	return fmt.Errorf("no PM agent found for project %s", projectID)
}
