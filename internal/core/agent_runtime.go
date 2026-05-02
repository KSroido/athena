package core

import (
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	"github.com/ksroido/athena/internal/db"
)

// AgentStatus represents the current status of an Agent
type AgentStatus string

const (
	StatusIdle      AgentStatus = "idle"
	StatusWorking   AgentStatus = "working"
	StatusInMeeting AgentStatus = "in_meeting"
	StatusOffline   AgentStatus = "offline"
)

// AgentProcess wraps a running Agent subprocess
type AgentProcess struct {
	ID           string
	Role         string
	ProjectID    string
	Cmd          *exec.Cmd
	Stdin        io.WriteCloser
	Stdout       io.ReadCloser
	Status       AgentStatus
	RestartCount int
	LastActive   time.Time
	cancel       context.CancelFunc
}

// AgentMessage is the message format sent to an Agent via stdin
type AgentMessage struct {
	Type      string `json:"type"`       // "task", "meeting_invite", "steer"
	TaskID    string `json:"task_id,omitempty"`
	Content   string `json:"content"`
	MeetingID string `json:"meeting_id,omitempty"`
	Agenda    string `json:"agenda,omitempty"`
}

// AgentResponse is the response format from an Agent via stdout
type AgentResponse struct {
	Type    string `json:"type"`     // "task_result", "tool_call", "blackboard_write", "meeting_message", "error"
	TaskID  string `json:"task_id,omitempty"`
	Content string `json:"content"`
	Error   string `json:"error,omitempty"`
}

// Supervisor manages all Agent subprocesses
type Supervisor struct {
	agents    map[string]*AgentProcess
	mu        sync.RWMutex
	agentBin  string // path to athena-agent binary
	llmConfig LLMConfig
	mainDB    *db.DB
}

// LLMConfig holds LLM configuration for agent processes
type LLMConfig struct {
	BaseURL string
	APIKey  string
	Model   string
}

// NewSupervisor creates a new Agent supervisor
func NewSupervisor(agentBin string, llmCfg LLMConfig, mainDB *db.DB) *Supervisor {
	return &Supervisor{
		agents:    make(map[string]*AgentProcess),
		agentBin:  agentBin,
		llmConfig: llmCfg,
		mainDB:    mainDB,
	}
}

// StartAgent launches a new Agent subprocess
func (s *Supervisor) StartAgent(ctx context.Context, agent *db.Agent, projectID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already running
	if _, exists := s.agents[agent.ID]; exists {
		return fmt.Errorf("agent %s is already running", agent.ID)
	}

	agentCtx, cancel := context.WithCancel(ctx)

	cmd := exec.CommandContext(agentCtx, s.agentBin,
		"--id", agent.ID,
		"--role", agent.Role,
		"--project", projectID,
	)

	// Set environment variables for LLM config
	cmd.Env = append(os.Environ(),
		"ATHENA_LLM_BASE_URL="+s.llmConfig.BaseURL,
		"ATHENA_LLM_API_KEY="+s.llmConfig.APIKey,
		"ATHENA_LLM_MODEL="+s.llmConfig.Model,
	)

	stdin, err := cmd.StdinPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("create stdin pipe: %w", err)
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		return fmt.Errorf("create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		return fmt.Errorf("start agent process: %w", err)
	}

	proc := &AgentProcess{
		ID:        agent.ID,
		Role:      agent.Role,
		ProjectID: projectID,
		Cmd:       cmd,
		Stdin:     stdin,
		Stdout:    stdout,
		Status:    StatusIdle,
		cancel:    cancel,
	}

	s.agents[agent.ID] = proc

	// Watch for agent process exit
	go s.watchAgent(agentCtx, proc)

	// Read agent stdout
	go s.listenAgent(agentCtx, proc)

	return nil
}

// StopAgent terminates an Agent subprocess
func (s *Supervisor) StopAgent(agentID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	proc, exists := s.agents[agentID]
	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	proc.cancel()
	delete(s.agents, agentID)
	return nil
}

// GetAgent returns information about a running agent
func (s *Supervisor) GetAgent(agentID string) (*AgentProcess, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	proc, exists := s.agents[agentID]
	return proc, exists
}

// ListAgents returns all running agents
func (s *Supervisor) ListAgents() []*AgentProcess {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make([]*AgentProcess, 0, len(s.agents))
	for _, proc := range s.agents {
		result = append(result, proc)
	}
	return result
}

// SendTask sends a task to an agent via stdin
func (s *Supervisor) SendTask(agentID string, taskID string, content string) error {
	s.mu.RLock()
	proc, exists := s.agents[agentID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	proc.Status = StatusWorking
	msg := AgentMessage{
		Type:    "task",
		TaskID:  taskID,
		Content: content,
	}

	return writeAgentMessage(proc.Stdin, msg)
}

// SendSteer sends a steering message (new requirements from CEO) to an agent
func (s *Supervisor) SendSteer(agentID string, content string) error {
	s.mu.RLock()
	proc, exists := s.agents[agentID]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("agent %s not found", agentID)
	}

	msg := AgentMessage{
		Type:    "steer",
		Content: content,
	}

	return writeAgentMessage(proc.Stdin, msg)
}

// watchAgent monitors an agent subprocess for crashes
func (s *Supervisor) watchAgent(ctx context.Context, proc *AgentProcess) {
	err := proc.Cmd.Wait()

	s.mu.Lock()
	defer s.mu.Unlock()

	if _, exists := s.agents[proc.ID]; !exists {
		return // already stopped intentionally
	}

	// Agent crashed — attempt restart (max 3 times)
	if proc.RestartCount < 3 && ctx.Err() == nil {
		proc.RestartCount++
		// Re-launch the agent
		delete(s.agents, proc.ID)
		_ = s.StartAgent(ctx, &db.Agent{
			ID:   proc.ID,
			Role: proc.Role,
		}, proc.ProjectID)
	} else {
		proc.Status = StatusOffline
		delete(s.agents, proc.ID)
	}

	_ = err
}

// listenAgent reads stdout from an agent process
func (s *Supervisor) listenAgent(ctx context.Context, proc *AgentProcess) {
	// Read and process agent responses
	// In a real implementation, this would decode JSON messages from stdout
	// and route them appropriately (blackboard writes, tool calls, etc.)
	//
	// For Phase 1, we just read and discard
	buf := make([]byte, 4096)
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, err := proc.Stdout.Read(buf)
			if err != nil {
				return
			}
		}
	}
}

// writeAgentMessage writes a JSON message to an agent's stdin
func writeAgentMessage(w io.Writer, msg AgentMessage) error {
	// In a real implementation, use json.NewEncoder(w).Encode(msg)
	// For now, placeholder
	return nil
}
