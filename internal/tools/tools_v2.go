package tools

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/db"
	"github.com/ksroido/athena/internal/hr"
)

// --- Assign Task Tool ---

// AssignTaskInput is the input for the assign_task tool
type AssignTaskInput struct {
	AgentRole string `json:"agent_role" jsonschema:"description=Target agent role (developer/tester/reviewer/designer),required"`
	TaskTitle string `json:"task_title" jsonschema:"description=Short task title,required"`
	TaskDesc  string `json:"task_desc" jsonschema:"description=Detailed task description,required"`
	Priority  int    `json:"priority" jsonschema:"description=Task priority 1-10. Default 5."`
}

// AssignTaskOutput is the output for the assign_task tool
type AssignTaskOutput struct {
	TaskID  string `json:"task_id"`
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
}

// NewAssignTaskTool creates a tool for PM to assign tasks to agents
// The taskFunc callback is used to actually send the task to the running agent
func NewAssignTaskTool(projectID, pmAgentID string, mainDB *db.DB, taskFunc func(agentID, taskID, content, fromAgent string) error, hireFunc func(req *hr.HireRequest) (*db.Agent, error)) (tool.InvokableTool, error) {
	return utils.InferTool(
		"assign_task",
		"Assign a task to an agent by role. If no agent with that role exists for the project, automatically request HR to hire one.",
		func(ctx context.Context, input AssignTaskInput) (*AssignTaskOutput, error) {
			priority := input.Priority
			if priority <= 0 {
				priority = 5
			}

			// Find agent with the specified role in this project
			var agentID string
			err := mainDB.DB().QueryRow(
				"SELECT a.id FROM agents a JOIN project_members pm ON a.id = pm.agent_id WHERE pm.project_id = ? AND a.role = ? AND a.status != 'offline' LIMIT 1",
				projectID, input.AgentRole,
			).Scan(&agentID)

			if err != nil {
				// No agent found — request HR to hire one
				if hireFunc != nil {
					agent, hireErr := hireFunc(&hr.HireRequest{
						Role:      input.AgentRole,
						ProjectID: projectID,
						Reason:    fmt.Sprintf("PM requested for task: %s", input.TaskTitle),
					})
					if hireErr != nil {
						return nil, fmt.Errorf("no %s agent found and HR hire failed: %v", input.AgentRole, hireErr)
					}
					agentID = agent.ID
				} else {
					return nil, fmt.Errorf("no %s agent found for project %s", input.AgentRole, projectID)
				}
			}

			// Create task record
			taskID := uuid.New().String()[:8]
			_, err = mainDB.DB().Exec(`
				INSERT INTO agent_tasks (id, project_id, agent_id, title, description, status, priority, created_at)
				VALUES (?, ?, ?, ?, ?, 'pending', ?, ?)
			`, taskID, projectID, agentID, input.TaskTitle, input.TaskDesc, priority, time.Now())
			if err != nil {
				return nil, fmt.Errorf("create task record: %w", err)
			}

			// Send task to the agent
			content := fmt.Sprintf("%s\n\n%s", input.TaskTitle, input.TaskDesc)
			if taskFunc != nil {
				if err := taskFunc(agentID, taskID, content, pmAgentID); err != nil {
					return &AssignTaskOutput{
						TaskID:  taskID,
						AgentID: agentID,
						Message: fmt.Sprintf("任务已创建 (ID: %s) 但发送失败: %v", taskID, err),
					}, nil
				}
			}

			// Update task status to in_progress
			mainDB.DB().Exec("UPDATE agent_tasks SET status = 'in_progress' WHERE id = ?", taskID)

			return &AssignTaskOutput{
				TaskID:  taskID,
				AgentID: agentID,
				Message: fmt.Sprintf("任务 '%s' 已分配给 %s Agent (%s)", input.TaskTitle, input.AgentRole, agentID),
			}, nil
		},
	)
}

// --- HR Request Tool ---

// HRRequestInput is the input for the hr_request tool
type HRRequestInput struct {
	Role   string `json:"role" jsonschema:"description=Role to hire (developer/tester/reviewer/designer),required"`
	Reason string `json:"reason" jsonschema:"description=Why this role is needed,required"`
}

// HRRequestOutput is the output for the hr_request tool
type HRRequestOutput struct {
	AgentID string `json:"agent_id"`
	Message string `json:"message"`
}

// NewHRRequestTool creates a tool for agents to request HR to hire new agents
func NewHRRequestTool(projectID string, hireFunc func(req *hr.HireRequest) (*db.Agent, error)) (tool.InvokableTool, error) {
	return utils.InferTool(
		"hr_request",
		"Request HR to hire a new agent for the project. Use when you need a role that doesn't exist yet.",
		func(ctx context.Context, input HRRequestInput) (*HRRequestOutput, error) {
			if hireFunc == nil {
				return nil, fmt.Errorf("HR function not available")
			}

			agent, err := hireFunc(&hr.HireRequest{
				Role:      input.Role,
				ProjectID: projectID,
				Reason:    input.Reason,
			})
			if err != nil {
				return nil, fmt.Errorf("HR hire failed: %v", err)
			}

			return &HRRequestOutput{
				AgentID: agent.ID,
				Message: fmt.Sprintf("HR已招聘 %s (%s)", agent.Name, agent.ID),
			}, nil
		},
	)
}

// --- File Read Tool ---

// FileReadInput is the input for the file_read tool
type FileReadInput struct {
	Path    string `json:"path" jsonschema:"description=File path to read (relative to project workspace),required"`
	Offset  int    `json:"offset" jsonschema:"description=Line number to start reading from (1-indexed). Default 1."`
	Limit   int    `json:"limit" jsonschema:"description=Maximum lines to read. Default 500."`
}

// FileReadOutput is the output for the file_read tool
type FileReadOutput struct {
	Content string `json:"content"`
	Lines   int    `json:"lines"`
	Path    string `json:"path"`
}

// NewFileReadTool creates a tool for reading files
func NewFileReadTool(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"file_read",
		"Read a file from the project workspace. Supports line offset and limit for large files.",
		func(ctx context.Context, input FileReadInput) (*FileReadOutput, error) {
			path := filepath.Join(workspaceDir, input.Path)
			path = filepath.Clean(path)

			// Security: ensure path is within workspace
			if !strings.HasPrefix(path, filepath.Clean(workspaceDir)) {
				return nil, fmt.Errorf("path outside workspace: %s", input.Path)
			}

			f, err := os.Open(path)
			if err != nil {
				return nil, fmt.Errorf("open file: %w", err)
			}
			defer f.Close()

			offset := input.Offset
			if offset <= 0 {
				offset = 1
			}
			limit := input.Limit
			if limit <= 0 {
				limit = 500
			}

			scanner := bufio.NewScanner(f)
			var lines []string
			lineNum := 0
			for scanner.Scan() {
				lineNum++
				if lineNum < offset {
					continue
				}
				if len(lines) >= limit {
					break
				}
				lines = append(lines, scanner.Text())
			}

			return &FileReadOutput{
				Content: strings.Join(lines, "\n"),
				Lines:   len(lines),
				Path:    input.Path,
			}, nil
		},
	)
}

// --- File Write Tool ---

// FileWriteInput is the input for the file_write tool
type FileWriteInput struct {
	Path    string `json:"path" jsonschema:"description=File path to write (relative to project workspace),required"`
	Content string `json:"content" jsonschema:"description=Content to write to the file,required"`
	Append  bool   `json:"append" jsonschema:"description=Append to file instead of overwriting. Default false."`
}

// FileWriteOutput is the output for the file_write tool
type FileWriteOutput struct {
	Path    string `json:"path"`
	Message string `json:"message"`
}

// NewFileWriteTool creates a tool for writing files
func NewFileWriteTool(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"file_write",
		"Write content to a file in the project workspace. Creates parent directories automatically.",
		func(ctx context.Context, input FileWriteInput) (*FileWriteOutput, error) {
			path := filepath.Join(workspaceDir, input.Path)
			path = filepath.Clean(path)

			// Security: ensure path is within workspace
			if !strings.HasPrefix(path, filepath.Clean(workspaceDir)) {
				return nil, fmt.Errorf("path outside workspace: %s", input.Path)
			}

			// Create parent directories
			if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
				return nil, fmt.Errorf("create directories: %w", err)
			}

			flag := os.O_WRONLY | os.O_CREATE
			if input.Append {
				flag |= os.O_APPEND
			} else {
				flag |= os.O_TRUNC
			}

			f, err := os.OpenFile(path, flag, 0644)
			if err != nil {
				return nil, fmt.Errorf("open file: %w", err)
			}
			defer f.Close()

			if _, err := f.WriteString(input.Content); err != nil {
				return nil, fmt.Errorf("write file: %w", err)
			}

			return &FileWriteOutput{
				Path:    input.Path,
				Message: "文件写入成功",
			}, nil
		},
	)
}

// --- Term Tool (actual execution version) ---

// TermExecInput is the input for the term tool
type TermExecInput struct {
	Command string `json:"command" jsonschema:"description=Shell command to execute,required"`
	WorkDir string `json:"workdir" jsonschema:"description=Working directory for the command. Default: project workspace."`
}

// TermExecOutput is the output for the term tool
type TermExecOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command"`
}

// NewTermExecTool creates a tool for executing shell commands (with basic safety)
func NewTermExecTool(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"term",
		"Execute a shell command in the project workspace. Dangerous commands (rm -rf /, dd, fork bombs) will be blocked.",
		func(ctx context.Context, input TermExecInput) (*TermExecOutput, error) {
			// Basic safety check (no LLM review for simplicity in Phase 2)
			cmd := strings.TrimSpace(input.Command)
			if isDangerousCommand(cmd) {
				return nil, fmt.Errorf("command blocked for safety: %s", cmd)
			}

			workDir := input.WorkDir
			if workDir == "" {
				workDir = workspaceDir
			}
			workDir = filepath.Join(workspaceDir, workDir)
			workDir = filepath.Clean(workDir)

			// Execute with bash
			execCmd := exec.CommandContext(ctx, "bash", "-c", cmd)
			execCmd.Dir = workDir

			stdout, err := execCmd.CombinedOutput()
			exitCode := 0
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					exitCode = exitErr.ExitCode()
				} else {
					return nil, fmt.Errorf("execute command: %w", err)
				}
			}

			return &TermExecOutput{
				Stdout:   string(stdout),
				ExitCode: exitCode,
				Command:  cmd,
			}, nil
		},
	)
}

// isDangerousCommand performs basic safety checks
func isDangerousCommand(cmd string) bool {
	dangerous := []string{
		"rm -rf /",
		"rm -rf /*",
		"dd if=/dev/zero",
		"dd if=/dev/random",
		":(){ :|:& };:",
		"chmod -R 777 /",
		"> /etc/passwd",
		"mkfs",
		"format",
	}
	lower := strings.ToLower(cmd)
	for _, d := range dangerous {
		if strings.Contains(lower, strings.ToLower(d)) {
			return true
		}
	}
	return false
}

// --- Memory Tool (file-based version) ---

// NewMemoryReadToolFile creates a memory read tool backed by a file
func NewMemoryReadToolFile(dataDir, agentID string) (tool.InvokableTool, error) {
	memoryPath := filepath.Join(dataDir, "agents", agentID, "memory.md")
	return utils.InferTool(
		"memory_read",
		"Read your personal memory (memory.md). Contains declarative facts separated by § symbols.",
		func(ctx context.Context, input MemoryReadInput) (*MemoryReadOutput, error) {
			data, err := os.ReadFile(memoryPath)
			if err != nil {
				if os.IsNotExist(err) {
					return &MemoryReadOutput{Content: ""}, nil
				}
				return nil, fmt.Errorf("read memory: %w", err)
			}
			return &MemoryReadOutput{Content: string(data)}, nil
		},
	)
}

// NewMemoryWriteToolFile creates a memory write tool backed by a file
func NewMemoryWriteToolFile(dataDir, agentID string) (tool.InvokableTool, error) {
	memoryPath := filepath.Join(dataDir, "agents", agentID, "memory.md")
	return utils.InferTool(
		"memory_write",
		"Append a declarative fact to your personal memory (memory.md). Facts are §-separated. Write only facts, not instructions.",
		func(ctx context.Context, input MemoryWriteInput) (*MemoryWriteOutput, error) {
			// Ensure directory exists
			if err := os.MkdirAll(filepath.Dir(memoryPath), 0755); err != nil {
				return nil, fmt.Errorf("create memory dir: %w", err)
			}

			// Append fact
			f, err := os.OpenFile(memoryPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("open memory: %w", err)
			}
			defer f.Close()

			fmt.Fprintf(f, "§ %s\n", input.Fact)

			return &MemoryWriteOutput{Message: "Fact saved to personal memory"}, nil
		},
	)
}

// Keep old function signatures for backward compatibility
// NewMemoryReadTool is deprecated — use NewMemoryReadToolFile
func NewMemoryReadTool(agentDataDir, agentID string) (tool.InvokableTool, error) {
	return NewMemoryReadToolFile(agentDataDir, agentID)
}

// NewMemoryWriteTool is deprecated — use NewMemoryWriteToolFile
func NewMemoryWriteTool(agentDataDir, agentID string) (tool.InvokableTool, error) {
	return NewMemoryWriteToolFile(agentDataDir, agentID)
}

// NewTermTool is deprecated — use NewTermExecTool
func NewTermTool(agentID string, safetyLLM func(ctx context.Context, prompt string) (string, error)) (tool.InvokableTool, error) {
	return NewTermExecTool(".")
}

// NewMeetingTool is a placeholder for Phase 3
func NewMeetingTool(agentID, projectID string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"meeting",
		"Participate in meetings: speak (add your input), resolve (propose a resolution), or close (end the meeting).",
		func(ctx context.Context, input MeetingInput) (*MeetingOutput, error) {
			return &MeetingOutput{
				Message: fmt.Sprintf("Meeting action '%s' received. Meeting system will be implemented in Phase 3.", input.Action),
			}, nil
		},
	)
}

// Ensure db import is used
var _ = (*db.BlackboardEntry)(nil)
var _ = (*blackboard.Board)(nil)

// --- Type definitions for legacy tools ---

// MemoryReadInput is the input for the memory read tool
type MemoryReadInput struct {
	// No input needed — reads the entire memory file
}

// MemoryReadOutput is the output for the memory read tool
type MemoryReadOutput struct {
	Content string `json:"content"`
}

// MemoryWriteInput is the input for the memory write tool
type MemoryWriteInput struct {
	Fact string `json:"fact" jsonschema:"description=A declarative fact to append to personal memory (§-separated),required"`
}

// MemoryWriteOutput is the output for the memory write tool
type MemoryWriteOutput struct {
	Message string `json:"message"`
}

// MeetingInput is the input for the meeting tool
type MeetingInput struct {
	Action    string `json:"action" jsonschema:"description=Action: speak/resolve/close,required"`
	MeetingID string `json:"meeting_id" jsonschema:"description=Meeting ID"`
	Content   string `json:"content" jsonschema:"description=Speech content (for speak) or resolution (for resolve)"`
}

// MeetingOutput is the output for the meeting tool
type MeetingOutput struct {
	Message string `json:"message"`
}
