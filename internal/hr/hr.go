package hr

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"

	"github.com/ksroido/athena/internal/db"
)

// AgentStarter is the interface for starting agents (injected by core)
type AgentStarter interface {
	StartAgentFromHR(agent *db.Agent, projectID string) error
}

// HR manages agent hiring and templates
type HR struct {
	mainDB  *db.DB
	starter AgentStarter
	dataDir string
}

// New creates a new HR instance
func New(mainDB *db.DB, starter AgentStarter, dataDir string) *HR {
	return &HR{
		mainDB:  mainDB,
		starter: starter,
		dataDir: dataDir,
	}
}

// RoleTemplate defines a template for creating agents
type RoleTemplate struct {
	Role        string
	Name        string
	Description string
	Tools       []string // tool names
}

// DefaultTemplates are the built-in role templates
var DefaultTemplates = map[string]RoleTemplate{
	"pm": {
		Role:        "pm",
		Name:        "项目经理",
		Description: "负责需求拆解、任务分配、验收交付",
		Tools:       []string{"blackboard_read", "blackboard_write", "assign_task", "hr_request", "memory_read", "memory_write"},
	},
	"developer": {
		Role:        "developer",
		Name:        "开发工程师",
		Description: "负责编写代码和实现功能",
		Tools:       []string{"blackboard_read", "blackboard_write", "term", "file_read", "file_write", "memory_read", "memory_write"},
	},
	"tester": {
		Role:        "tester",
		Name:        "测试工程师",
		Description: "负责编写测试用例和执行测试",
		Tools:       []string{"blackboard_read", "blackboard_write", "term", "file_read", "file_write", "memory_read", "memory_write"},
	},
	"reviewer": {
		Role:        "reviewer",
		Name:        "代码审查员",
		Description: "负责代码审查，上下文与开发隔离",
		Tools:       []string{"blackboard_read", "blackboard_write", "file_read", "memory_read", "memory_write"},
	},
	"designer": {
		Role:        "designer",
		Name:        "设计师",
		Description: "负责UI/UX设计和视觉规范",
		Tools:       []string{"blackboard_read", "blackboard_write", "file_read", "file_write", "memory_read", "memory_write"},
	},
}

// HireRequest is a request to hire a new agent
type HireRequest struct {
	Role      string `json:"role"`
	ProjectID string `json:"project_id"`
	Reason    string `json:"reason"`
}

// Hire creates and starts a new agent based on a role template
func (h *HR) Hire(req *HireRequest) (*db.Agent, error) {
	// 1. Check company size limit
	maxAgents := 100
	var count int
	h.mainDB.DB().QueryRow("SELECT COUNT(*) FROM agents WHERE status != 'offline'").Scan(&count)
	if count >= maxAgents {
		return nil, fmt.Errorf("公司人数已达上限 (%d/%d)，请联系CEO扩容", count, maxAgents)
	}

	// 2. Look up template
	tmpl, ok := DefaultTemplates[req.Role]
	if !ok {
		return nil, fmt.Errorf("未知角色模板: %s，可用: pm, developer, tester, reviewer, designer", req.Role)
	}

	// 3. Check if this project already has an agent with this role
	var existingCount int
	h.mainDB.DB().QueryRow(
		"SELECT COUNT(*) FROM agents a JOIN project_members pm ON a.id = pm.agent_id WHERE pm.project_id = ? AND a.role = ? AND a.status != 'offline'",
		req.ProjectID, req.Role,
	).Scan(&existingCount)
	if existingCount > 0 {
		return nil, fmt.Errorf("项目 %s 已有 %s 角色的 Agent", req.ProjectID, req.Role)
	}

	// 4. Create agent record
	agentID := fmt.Sprintf("%s-%s-%s", req.ProjectID, req.Role, uuid.New().String()[:8])
	agent := &db.Agent{
		ID:        agentID,
		Name:      tmpl.Name,
		Role:      tmpl.Role,
		Status:    "idle",
		Tools:     toolsToJSON(tmpl.Tools),
		Model:     "default",
		CreatedBy: "hr",
		CreatedAt: time.Now(),
	}

	_, err := h.mainDB.DB().Exec(`
		INSERT INTO agents (id, name, role, status, tools, model, created_by, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, agent.ID, agent.Name, agent.Role, agent.Status, agent.Tools, agent.Model, agent.CreatedBy, agent.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("insert agent: %w", err)
	}

	// 5. Add to project members
	_, err = h.mainDB.DB().Exec(`
		INSERT INTO project_members (id, project_id, agent_id, role, joined_at)
		VALUES (?, ?, ?, ?, ?)
	`, uuid.New().String()[:8], req.ProjectID, agent.ID, agent.Role, time.Now())
	if err != nil {
		return nil, fmt.Errorf("add project member: %w", err)
	}

	// 6. Create agent data directory
	agentDir := filepath.Join(h.dataDir, "agents", agent.ID)
	if err := os.MkdirAll(agentDir, 0755); err != nil {
		log.Printf("[hr] warning: failed to create agent dir %s: %v", agentDir, err)
	}

	// 7. Initialize memory.md
	memoryPath := filepath.Join(agentDir, "memory.md")
	_ = os.WriteFile(memoryPath, []byte(fmt.Sprintf("# %s 个人记忆\n\n", agent.Name)), 0644)

	// 8. Start agent goroutine
	if err := h.starter.StartAgentFromHR(agent, req.ProjectID); err != nil {
		log.Printf("[hr] failed to start agent %s: %v", agent.ID, err)
	}

	log.Printf("[hr] hired %s (%s) for project %s", agent.Name, agent.ID, req.ProjectID)
	return agent, nil
}

// ListCompany returns all agents
func (h *HR) ListCompany() ([]*db.Agent, error) {
	rows, err := h.mainDB.DB().Query(
		"SELECT id, name, role, status, model, created_by, created_at FROM agents ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var agents []*db.Agent
	for rows.Next() {
		a := &db.Agent{}
		if err := rows.Scan(&a.ID, &a.Name, &a.Role, &a.Status, &a.Model, &a.CreatedBy, &a.CreatedAt); err != nil {
			continue
		}
		agents = append(agents, a)
	}
	return agents, nil
}

// Fire removes an agent
func (h *HR) Fire(agentID string) error {
	_, err := h.mainDB.DB().Exec("UPDATE agents SET status = 'offline' WHERE id = ?", agentID)
	return err
}

// toolsToJSON converts a tool list to JSON string
func toolsToJSON(tools []string) string {
	if len(tools) == 0 {
		return "[]"
	}
	result := "["
	for i, t := range tools {
		if i > 0 {
			result += ","
		}
		result += fmt.Sprintf("\"%s\"", t)
	}
	result += "]"
	return result
}
