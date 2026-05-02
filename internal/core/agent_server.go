package core

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/db"
	"github.com/ksroido/athena/internal/hr"
)

// AgentServer is the CEO Secretary — the primary interface between the CEO and the agent system
type AgentServer struct {
	llm        *LLMClient
	mainDB     *db.DB
	manager    *AgentManager
	hr         *hr.HR
	dataDir    string
}

// NewAgentServer creates a new AgentServer
func NewAgentServer(llm *LLMClient, mainDB *db.DB, manager *AgentManager, hrInst *hr.HR, dataDir string) *AgentServer {
	return &AgentServer{
		llm:     llm,
		mainDB:  mainDB,
		manager: manager,
		hr:      hrInst,
		dataDir: dataDir,
	}
}

// IntentType represents the recognized intent of a CEO message
type IntentType string

const (
	IntentNewProject    IntentType = "new_project"
	IntentUpdateProject IntentType = "update_project"
	IntentQueryProject  IntentType = "query_project"
	IntentHR            IntentType = "hr_request"
	IntentGeneral       IntentType = "general"
)

// RecognizedIntent holds the result of intent recognition
type RecognizedIntent struct {
	Intent    IntentType `json:"intent"`
	ProjectID string     `json:"project_id,omitempty"`
	Content   string     `json:"content"`
}

// ProcessCEOMessage is the main entry point for CEO input
func (as *AgentServer) ProcessCEOMessage(ctx context.Context, message string) (string, error) {
	// Step 1: Recognize intent using LLM
	intent, err := as.recognizeIntent(ctx, message)
	if err != nil {
		return "", fmt.Errorf("recognize intent: %w", err)
	}

	log.Printf("[agent-server] intent=%s project=%s", intent.Intent, intent.ProjectID)

	// Step 2: Route based on intent
	switch intent.Intent {
	case IntentNewProject:
		return as.handleNewProject(ctx, message)
	case IntentUpdateProject:
		return as.handleUpdateProject(ctx, intent.ProjectID, message)
	case IntentQueryProject:
		return as.handleQueryProject(ctx, intent.ProjectID)
	case IntentHR:
		return as.handleHRRequest(ctx, message)
	default:
		return as.handleGeneral(ctx, message)
	}
}

// recognizeIntent uses LLM to identify the CEO's intent and match projects
func (as *AgentServer) recognizeIntent(ctx context.Context, message string) (*RecognizedIntent, error) {
	projects, err := as.listProjects()
	if err != nil {
		return nil, err
	}

	projectList := make([]string, len(projects))
	for i, p := range projects {
		projectList[i] = fmt.Sprintf("- UUID: %s, Name: %s, Status: %s, Requirement: %s",
			p.ID, p.Name, p.Status, truncate(p.OriginalRequirement, 100))
	}

	prompt := fmt.Sprintf(`你是一个意图识别系统。根据CEO的消息，识别意图类型和可能关联的项目。

意图类型：
- new_project: CEO提出了新项目需求
- update_project: CEO对已有项目提出新要求或修改
- query_project: CEO询问项目进度或状态
- hr_request: CEO提出人力资源相关要求（招人、裁员等）
- general: 其他一般性消息

当前项目列表：
%s

CEO消息: %s

请以JSON格式返回：{"intent": "意图类型", "project_id": "项目UUID(如有关联)", "content": "消息内容摘要"}`,
		strings.Join(projectList, "\n"), message)

	resp, err := as.llm.ChatWithSystem(ctx, "你是意图识别助手，只返回JSON。", prompt)
	if err != nil {
		return &RecognizedIntent{Intent: IntentGeneral, Content: message}, nil
	}

	var intent RecognizedIntent
	content := resp.Content
	start := strings.Index(content, "{")
	end := strings.LastIndex(content, "}")
	if start >= 0 && end > start {
		if err := json.Unmarshal([]byte(content[start:end+1]), &intent); err != nil {
			return &RecognizedIntent{Intent: IntentNewProject, Content: message}, nil
		}
	} else {
		return &RecognizedIntent{Intent: IntentNewProject, Content: message}, nil
	}

	return &intent, nil
}

// handleNewProject creates a new project, hires PM, and starts the workflow
func (as *AgentServer) handleNewProject(ctx context.Context, requirement string) (string, error) {
	// Generate project UUID
	projectID := generateUUID()

	// Create project in database
	_, err := as.mainDB.DB().Exec(`
		INSERT INTO projects (id, name, original_requirement, status)
		VALUES (?, ?, ?, 'active')
	`, projectID, extractProjectName(requirement), requirement)
	if err != nil {
		return "", fmt.Errorf("create project: %w", err)
	}

	// Open blackboard for this project
	board, err := blackboard.OpenBoard(as.mainDB.DataDir(), projectID)
	if err != nil {
		return "", fmt.Errorf("open blackboard: %w", err)
	}

	// Write CEO's original requirement as the first blackboard entry
	board.WriteEntrySync(&db.BlackboardEntry{
		ID:        generateUUID(),
		ProjectID: projectID,
		Category:  blackboard.CategoryGoal,
		Content:   requirement,
		Certainty: blackboard.CertaintyCertain,
		Author:    "ceo",
	})
	board.Close()

	// Create workspace directory for the project
	// workspaceDir is created on-demand by FileWriteTool

	// Hire PM Agent for this project
	pmAgent, err := as.hr.Hire(&hr.HireRequest{
		Role:      "pm",
		ProjectID: projectID,
		Reason:    "项目创建，需要项目经理拆解需求",
	})
	if err != nil {
		return fmt.Sprintf("项目已创建 (UUID: %s)，但PM招聘失败: %v", projectID, err), nil
	}

	// Send the requirement as a task to the PM
	taskID := generateUUID()
	if err := as.manager.SendTask(pmAgent.ID, taskID, requirement, "ceo"); err != nil {
		return fmt.Sprintf("项目已创建 (UUID: %s)，PM已招聘(%s)，但任务发送失败: %v", projectID, pmAgent.ID, err), nil
	}

	return fmt.Sprintf("项目已创建 (UUID: %s)。PM(%s)已接收需求，正在分析和拆解...", projectID, pmAgent.ID), nil
}

// handleUpdateProject forwards an update to the project's PM Agent (steer mode)
func (as *AgentServer) handleUpdateProject(ctx context.Context, projectID string, update string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("无法匹配到相关项目，请提供更具体的项目描述")
	}

	// Write the update to the project's blackboard
	board, err := blackboard.OpenBoard(as.mainDB.DataDir(), projectID)
	if err != nil {
		return "", fmt.Errorf("open blackboard: %w", err)
	}
	defer board.Close()

	board.WriteEntrySync(&db.BlackboardEntry{
		ID:        generateUUID(),
		ProjectID: projectID,
		Category:  blackboard.CategoryGoal,
		Content:   fmt.Sprintf("[CEO新需求] %s", update),
		Certainty: blackboard.CertaintyCertain,
		Author:    "ceo",
	})

	// Find the PM agent for this project
	var pmAgentID string
	err = as.mainDB.DB().QueryRow(
		"SELECT a.id FROM agents a JOIN project_members pm ON a.id = pm.agent_id WHERE pm.project_id = ? AND a.role = 'pm' AND a.status != 'offline' LIMIT 1",
		projectID,
	).Scan(&pmAgentID)

	if err == nil {
		// Send steer to PM Agent
		err = as.manager.SendSteer(pmAgentID, update)
		if err != nil {
			return fmt.Sprintf("需求已写入项目 %s 的黑板，但PM通知失败", projectID), nil
		}
		return fmt.Sprintf("新需求已传递给项目 %s 的项目经理（并行模式，不影响当前执行）", projectID), nil
	}

	return fmt.Sprintf("需求已写入项目 %s 的黑板，等待项目经理处理", projectID), nil
}

// handleQueryProject returns the status of a project
func (as *AgentServer) handleQueryProject(ctx context.Context, projectID string) (string, error) {
	if projectID == "" {
		return "", fmt.Errorf("无法匹配到相关项目")
	}

	var project db.Project
	err := as.mainDB.DB().QueryRow(
		"SELECT id, name, status, original_requirement FROM projects WHERE id = ?",
		projectID,
	).Scan(&project.ID, &project.Name, &project.Status, &project.OriginalRequirement)
	if err == sql.ErrNoRows {
		return "", fmt.Errorf("项目 %s 不存在", projectID)
	} else if err != nil {
		return "", fmt.Errorf("query project: %w", err)
	}

	// Read blackboard entries for this project
	board, err := blackboard.OpenBoard(as.mainDB.DataDir(), projectID)
	if err != nil {
		return "", fmt.Errorf("open blackboard: %w", err)
	}
	defer board.Close()

	entries, _ := board.ReadEntries("", 50, 0)

	// Count agents
	var agentCount int
	as.mainDB.DB().QueryRow(
		"SELECT COUNT(*) FROM agents a JOIN project_members pm ON a.id = pm.agent_id WHERE pm.project_id = ? AND a.status != 'offline'",
		projectID,
	).Scan(&agentCount)

	// List running agents
	runningAgents := as.manager.ListAgents()
	var projectAgents []string
	for _, a := range runningAgents {
		if a.ProjectID == projectID {
			projectAgents = append(projectAgents, fmt.Sprintf("%s(%s)", a.Role, a.Status))
		}
	}

	agentInfo := strings.Join(projectAgents, ", ")
	if agentInfo == "" {
		agentInfo = "无运行中的Agent"
	}

	result := fmt.Sprintf("项目: %s (UUID: %s)\n状态: %s\n原始需求: %s\n黑板条目数: %d\n团队人数: %d\n运行中: %s",
		project.Name, project.ID, project.Status, truncate(project.OriginalRequirement, 100), len(entries), agentCount, agentInfo)

	return result, nil
}

// handleHRRequest handles HR-related CEO requests
func (as *AgentServer) handleHRRequest(ctx context.Context, message string) (string, error) {
	// List current company
	agents, err := as.hr.ListCompany()
	if err != nil {
		return "", err
	}

	var agentList []string
	for _, a := range agents {
		agentList = append(agentList, fmt.Sprintf("- %s (%s, 角色: %s, 状态: %s)", a.Name, a.ID, a.Role, a.Status))
	}

	return fmt.Sprintf("当前公司成员 (%d人):\n%s\n\n请告诉我要招聘什么角色，或者对人员做什么调整。", len(agents), strings.Join(agentList, "\n")), nil
}

// handleGeneral handles general CEO messages
func (as *AgentServer) handleGeneral(ctx context.Context, message string) (string, error) {
	resp, err := as.llm.ChatWithSystem(ctx,
		"你是Athena系统的CEO秘书。简洁回答CEO的问题。",
		message,
	)
	if err != nil {
		return "", err
	}
	return resp.Content, nil
}

// listProjects returns all projects from the database
func (as *AgentServer) listProjects() ([]*db.Project, error) {
	rows, err := as.mainDB.DB().Query(
		"SELECT id, name, status, original_requirement, requirement_summary FROM projects ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var projects []*db.Project
	for rows.Next() {
		p := &db.Project{}
		if err := rows.Scan(&p.ID, &p.Name, &p.Status, &p.OriginalRequirement, &p.RequirementSummary); err != nil {
			continue
		}
		projects = append(projects, p)
	}
	return projects, nil
}

// extractProjectName tries to extract a project name from the requirement
func extractProjectName(requirement string) string {
	requirement = strings.TrimSpace(requirement)
	if len(requirement) > 50 {
		return requirement[:50] + "..."
	}
	return requirement
}

// truncate truncates a string to maxLen
func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// Manager returns the AgentManager
func (as *AgentServer) Manager() *AgentManager {
	return as.manager
}

// HR returns the HR instance
func (as *AgentServer) HR() *hr.HR {
	return as.hr
}
