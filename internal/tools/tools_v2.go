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
	"unicode/utf8"

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
	AgentRole string `json:"agent_role" jsonschema:"description=Target agent role ID (e.g. dev.frontend, dev.backend, dev.backend.finance, tester, reviewer, designer). Matches the role used in hr_request.,required"`
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
	Role       string `json:"role" jsonschema:"description=Role ID to hire. Can be a registered role (e.g. dev.frontend, dev.backend, tester, reviewer, designer) or any custom role ID (e.g. dev.backend.finance, dev.backend.security, tester.security). HR will auto-generate a professional soul for custom roles.,required"`
	Speciality string `json:"speciality" jsonschema:"description=Speciality hint for LLM soul generation. E.g. '金融量化交易系统开发', '渗透测试与漏洞扫描', '品牌视觉与CSS动效'. Optional for registered roles, required for custom roles to ensure soul quality."`
	Reason     string `json:"reason" jsonschema:"description=Why this role is needed for the project,required"`
}

// HRRequestOutput is the output for the hr_request tool
type HRRequestOutput struct {
	AgentID string `json:"agent_id"`
	Role    string `json:"role"`
	Name    string `json:"name"`
	Message string `json:"message"`
}

// NewHRRequestTool creates a tool for agents to request HR to hire new agents.
// Supports both registered roles and dynamic custom roles.
func NewHRRequestTool(projectID string, hireFunc func(req *hr.HireRequest) (*db.Agent, error)) (tool.InvokableTool, error) {
	return utils.InferTool(
		"hr_request",
		"Request HR to hire a new agent for the project. You can use registered role IDs (dev.frontend, dev.backend, tester, etc.) OR specify any custom role ID (e.g. dev.backend.finance, dev.backend.security, tester.security). For custom roles, provide a speciality description so HR can generate a professional soul. HR will automatically create and save the role definition for future reuse.",
		func(ctx context.Context, input HRRequestInput) (*HRRequestOutput, error) {
			if hireFunc == nil {
				return nil, fmt.Errorf("HR function not available")
			}

			agent, err := hireFunc(&hr.HireRequest{
				Role:       input.Role,
				Speciality: input.Speciality,
				ProjectID:  projectID,
				Reason:     input.Reason,
			})
			if err != nil {
				return nil, fmt.Errorf("HR hire failed: %v", err)
			}

			return &HRRequestOutput{
				AgentID: agent.ID,
				Role:    agent.Role,
				Name:    agent.Name,
				Message: fmt.Sprintf("HR已招聘 %s (%s, 角色: %s)", agent.Name, agent.ID, agent.Role),
			}, nil
		},
	)
}

// --- File Read Tool ---

// FileReadInput is the input for the file_read tool
type FileReadInput struct {
	Path   string `json:"path" jsonschema:"description=File path to read (relative to project workspace),required"`
	Offset int    `json:"offset" jsonschema:"description=Line number to start reading from (1-indexed). Default 1."`
	Limit  int    `json:"limit" jsonschema:"description=Maximum lines to read. Default 500."`
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
		"Read a text file from the project workspace. Path must be relative to the workspace or an absolute path inside it. Returns selected lines plus the resolved workspace path for verification evidence.",
		func(ctx context.Context, input FileReadInput) (*FileReadOutput, error) {
			path, err := resolveWorkspacePath(workspaceDir, input.Path, false)
			if err != nil {
				return nil, err
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
		"Write content to a file in the project workspace. Path must be relative to the workspace or an absolute path inside it. Creates parent directories automatically; use append=true only for append-only logs.",
		func(ctx context.Context, input FileWriteInput) (*FileWriteOutput, error) {
			path, err := resolveWorkspacePath(workspaceDir, input.Path, true)
			if err != nil {
				return nil, err
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

func resolveWorkspacePath(workspaceDir, inputPath string, allowCreate bool) (string, error) {
	workspaceAbs, err := filepath.Abs(workspaceDir)
	if err != nil {
		return "", fmt.Errorf("resolve workspace: %w", err)
	}
	workspaceAbs, err = filepath.EvalSymlinks(workspaceAbs)
	if err != nil {
		return "", fmt.Errorf("resolve workspace symlinks: %w", err)
	}

	cleanInput := strings.TrimSpace(inputPath)
	if cleanInput == "" {
		return "", fmt.Errorf("path is required")
	}

	var candidate string
	if filepath.IsAbs(cleanInput) {
		candidate = filepath.Clean(cleanInput)
	} else {
		candidate = filepath.Join(workspaceAbs, cleanInput)
	}
	candidate, err = filepath.Abs(filepath.Clean(candidate))
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	if !pathInsideWorkspace(workspaceAbs, candidate) {
		return "", fmt.Errorf("path outside workspace: %s", inputPath)
	}

	if allowCreate {
		parent := filepath.Dir(candidate)
		if err := os.MkdirAll(parent, 0755); err != nil {
			return "", fmt.Errorf("create directories: %w", err)
		}
		realParent, err := filepath.EvalSymlinks(parent)
		if err != nil {
			return "", fmt.Errorf("resolve parent symlinks: %w", err)
		}
		if !pathInsideWorkspace(workspaceAbs, realParent) {
			return "", fmt.Errorf("path outside workspace after symlink resolution: %s", inputPath)
		}
		return filepath.Join(realParent, filepath.Base(candidate)), nil
	}

	realCandidate, err := filepath.EvalSymlinks(candidate)
	if err != nil {
		return "", fmt.Errorf("resolve path symlinks: %w", err)
	}
	if !pathInsideWorkspace(workspaceAbs, realCandidate) {
		return "", fmt.Errorf("path outside workspace after symlink resolution: %s", inputPath)
	}
	return realCandidate, nil
}

// --- Term Tool (actual execution version) ---

// TermExecInput is the input for the term tool
type TermExecInput struct {
	Command string `json:"command" jsonschema:"description=Shell command to execute,required"`
	WorkDir string `json:"workdir" jsonschema:"description=Working directory for the command. Default: project workspace."`
	Timeout int    `json:"timeout" jsonschema:"description=Maximum execution time in seconds. Default 120, maximum 600."`
}

// TermExecOutput is the output for the term tool
type TermExecOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	Command  string `json:"command"`
	WorkDir  string `json:"workdir"`
	TimedOut bool   `json:"timed_out"`
}

// NewTermExecTool creates a tool for executing shell commands (with basic safety)
func NewTermExecTool(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"term",
		"Execute a shell command in the project workspace. Supports relative workdir, timeout seconds, network/API calls, Python scripts, pip installs, tests, and builds. Dangerous commands (rm -rf /, dd, fork bombs) will be blocked.",
		func(ctx context.Context, input TermExecInput) (*TermExecOutput, error) {
			return runShellCommand(ctx, workspaceDir, input.Command, input.WorkDir, input.Timeout)
		},
	)
}

func runShellCommand(ctx context.Context, workspaceDir, command, workDir string, timeout int) (*TermExecOutput, error) {
	// Basic safety check (no LLM review for simplicity in Phase 2)
	cmd := strings.TrimSpace(command)
	if isDangerousCommand(cmd) {
		return nil, fmt.Errorf("command blocked for safety: %s", cmd)
	}

	workspaceAbs, err := filepath.Abs(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace: %w", err)
	}
	workspaceAbs, err = filepath.EvalSymlinks(workspaceAbs)
	if err != nil {
		return nil, fmt.Errorf("resolve workspace symlinks: %w", err)
	}

	resolvedWorkDir := workDir
	if resolvedWorkDir == "" || resolvedWorkDir == "." {
		resolvedWorkDir = workspaceAbs
	} else if filepath.IsAbs(resolvedWorkDir) {
		resolvedWorkDir = filepath.Clean(resolvedWorkDir)
	} else {
		resolvedWorkDir = filepath.Join(workspaceAbs, resolvedWorkDir)
	}
	resolvedWorkDir, err = filepath.Abs(filepath.Clean(resolvedWorkDir))
	if err != nil {
		return nil, fmt.Errorf("resolve workdir: %w", err)
	}

	if !pathInsideWorkspace(workspaceAbs, resolvedWorkDir) {
		return nil, fmt.Errorf("workdir outside workspace: %s", workDir)
	}
	if err := os.MkdirAll(resolvedWorkDir, 0755); err != nil {
		return nil, fmt.Errorf("create workdir: %w", err)
	}
	realWorkDir, err := filepath.EvalSymlinks(resolvedWorkDir)
	if err != nil {
		return nil, fmt.Errorf("resolve workdir symlinks: %w", err)
	}
	if !pathInsideWorkspace(workspaceAbs, realWorkDir) {
		return nil, fmt.Errorf("workdir outside workspace after symlink resolution: %s", workDir)
	}

	if timeout <= 0 {
		timeout = 120
	}
	if timeout > 600 {
		timeout = 600
	}
	cmdCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	// Execute with bash
	execCmd := exec.CommandContext(cmdCtx, "bash", "-c", cmd)
	execCmd.Dir = realWorkDir
	setProcessGroup(execCmd)

	stdout, err := execCmd.CombinedOutput()
	exitCode := 0
	if cmdCtx.Err() == context.DeadlineExceeded {
		if execCmd.Process != nil {
			killProcessGroup(execCmd.Process.Pid)
		}
		return &TermExecOutput{
			Stdout:   truncateOutput(string(stdout)+fmt.Sprintf("\n[timeout after %d seconds]", timeout), 12000),
			ExitCode: -1,
			Command:  cmd,
			WorkDir:  realWorkDir,
			TimedOut: true,
		}, nil
	}
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			exitCode = exitErr.ExitCode()
		} else {
			return nil, fmt.Errorf("execute command: %w", err)
		}
	}

	return &TermExecOutput{
		Stdout:   truncateOutput(string(stdout), 12000),
		ExitCode: exitCode,
		Command:  cmd,
		WorkDir:  realWorkDir,
	}, nil
}

func pathInsideWorkspace(workspaceAbs, candidate string) bool {
	rel, err := filepath.Rel(workspaceAbs, candidate)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator)) && !filepath.IsAbs(rel)
}

func truncateOutput(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	end := maxLen
	for end > 0 && !utf8.ValidString(s[:end]) {
		end--
	}
	return s[:end] + fmt.Sprintf("\n... [truncated, total %d bytes]", len(s))
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

// --- Python and dynamic tool primitives ---

type PythonExecInput struct {
	Code    string `json:"code" jsonschema:"description=Python code to execute from stdin,required"`
	WorkDir string `json:"workdir" jsonschema:"description=Working directory relative to project workspace. Default: workspace root."`
	Timeout int    `json:"timeout" jsonschema:"description=Maximum execution time in seconds. Default 120, maximum 600."`
}

type PythonExecOutput struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
	WorkDir  string `json:"workdir"`
	TimedOut bool   `json:"timed_out"`
}

func NewPythonExecTool(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"python",
		"Execute Python code provided directly in the code field. This is the built-in Python input interface for calculations, data processing, quick probes, and generating files inside the workspace. Use file_write for reusable scripts and tool_create_python for reusable dynamic tools.",
		func(ctx context.Context, input PythonExecInput) (*PythonExecOutput, error) {
			if strings.TrimSpace(input.Code) == "" {
				return nil, fmt.Errorf("python code is required")
			}
			cmd := "python3 - <<'PY'\n" + input.Code + "\nPY"
			out, err := runShellCommand(ctx, workspaceDir, cmd, input.WorkDir, input.Timeout)
			if err != nil {
				return nil, err
			}
			return &PythonExecOutput{Stdout: out.Stdout, Stderr: out.Stderr, ExitCode: out.ExitCode, WorkDir: out.WorkDir, TimedOut: out.TimedOut}, nil
		},
	)
}

type PythonToolCreateInput struct {
	Name        string `json:"name" jsonschema:"description=Dynamic tool name. Use lowercase letters, numbers, underscore, or dash only.,required"`
	Description string `json:"description" jsonschema:"description=What this tool does, when to use it, and expected JSON input.,required"`
	Code        string `json:"code" jsonschema:"description=Python script code. It receives JSON from argv[1] and must print JSON or text to stdout.,required"`
}

type PythonToolCreateOutput struct {
	Name        string `json:"name"`
	Path        string `json:"path"`
	Description string `json:"description"`
	Message     string `json:"message"`
}

func NewPythonToolCreateTool(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"tool_create_python",
		"Create or update a reusable dynamic Python tool under .athena/tools inside the workspace. The tool can later be executed with dynamic_python_tool by name. Use this when a missing capability should become reusable instead of one-off code.",
		func(ctx context.Context, input PythonToolCreateInput) (*PythonToolCreateOutput, error) {
			name, err := sanitizeDynamicToolName(input.Name)
			if err != nil {
				return nil, err
			}
			if strings.TrimSpace(input.Description) == "" {
				return nil, fmt.Errorf("description is required")
			}
			if strings.TrimSpace(input.Code) == "" {
				return nil, fmt.Errorf("code is required")
			}
			toolsDir, err := resolveWorkspacePath(workspaceDir, filepath.Join(".athena", "tools"), true)
			if err != nil {
				return nil, err
			}
			if err := os.MkdirAll(toolsDir, 0755); err != nil {
				return nil, fmt.Errorf("create tools dir: %w", err)
			}
			toolPath := filepath.Join(toolsDir, name+".py")
			metaPath := filepath.Join(toolsDir, name+".md")
			code := input.Code
			if !strings.HasPrefix(code, "#!") {
				code = "#!/usr/bin/env python3\n" + code
			}
			if err := os.WriteFile(toolPath, []byte(code), 0755); err != nil {
				return nil, fmt.Errorf("write dynamic python tool: %w", err)
			}
			meta := fmt.Sprintf("# %s\n\n%s\n", name, strings.TrimSpace(input.Description))
			if err := os.WriteFile(metaPath, []byte(meta), 0644); err != nil {
				return nil, fmt.Errorf("write dynamic python tool metadata: %w", err)
			}
			return &PythonToolCreateOutput{Name: name, Path: toolPath, Description: input.Description, Message: "dynamic Python tool saved; it will appear in refreshed tool inventory and can be run with dynamic_python_tool"}, nil
		},
	)
}

type DynamicPythonToolInput struct {
	Name    string `json:"name" jsonschema:"description=Dynamic Python tool name created by tool_create_python,required"`
	Input   string `json:"input" jsonschema:"description=JSON string or plain text passed as argv[1] to the dynamic tool."`
	Timeout int    `json:"timeout" jsonschema:"description=Maximum execution time in seconds. Default 120, maximum 600."`
}

type DynamicPythonToolOutput struct {
	Name     string `json:"name"`
	Stdout   string `json:"stdout"`
	ExitCode int    `json:"exit_code"`
	TimedOut bool   `json:"timed_out"`
}

func NewDynamicPythonToolRunner(workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"dynamic_python_tool",
		"Run a reusable dynamic Python tool created under .athena/tools. Pass tool-specific input as a JSON string in input. Use this after tool_create_python or when current tool inventory lists reusable Python tools.",
		func(ctx context.Context, input DynamicPythonToolInput) (*DynamicPythonToolOutput, error) {
			name, err := sanitizeDynamicToolName(input.Name)
			if err != nil {
				return nil, err
			}
			toolPath, err := resolveWorkspacePath(workspaceDir, filepath.Join(".athena", "tools", name+".py"), false)
			if err != nil {
				return nil, fmt.Errorf("dynamic tool not found: %w", err)
			}
			cmd := fmt.Sprintf("python3 %q %q", toolPath, input.Input)
			out, err := runShellCommand(ctx, workspaceDir, cmd, ".", input.Timeout)
			if err != nil {
				return nil, err
			}
			return &DynamicPythonToolOutput{Name: name, Stdout: out.Stdout, ExitCode: out.ExitCode, TimedOut: out.TimedOut}, nil
		},
	)
}

func sanitizeDynamicToolName(name string) (string, error) {
	clean := strings.TrimSpace(name)
	if clean == "" {
		return "", fmt.Errorf("tool name is required")
	}
	for _, r := range clean {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' || r == '-' {
			continue
		}
		return "", fmt.Errorf("invalid tool name %q: use lowercase letters, numbers, underscore, or dash", name)
	}
	return clean, nil
}

func DynamicPythonToolInventory(workspaceDir string) string {
	workspaceAbs, err := filepath.Abs(workspaceDir)
	if err != nil {
		return ""
	}
	workspaceAbs, err = filepath.EvalSymlinks(workspaceAbs)
	if err != nil {
		return ""
	}
	toolsDir := filepath.Join(workspaceAbs, ".athena", "tools")
	entries, err := os.ReadDir(toolsDir)
	if err != nil {
		return ""
	}
	var sb strings.Builder
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".py") {
			continue
		}
		name := strings.TrimSuffix(entry.Name(), ".py")
		desc := "workspace dynamic Python tool"
		if meta, err := os.ReadFile(filepath.Join(toolsDir, name+".md")); err == nil {
			desc = strings.TrimSpace(string(meta))
		}
		sb.WriteString(fmt.Sprintf("- `%s`: %s\n", name, strings.ReplaceAll(desc, "\n", " ")))
	}
	return sb.String()
}

// --- Self improvement tools ---

type SelfAssessInput struct {
	Focus string `json:"focus" jsonschema:"description=What capability, failure, or task aspect to assess. Optional."`
}

type SelfAssessOutput struct {
	SoulPath      string `json:"soul_path"`
	MemoryPath    string `json:"memory_path"`
	ToolDir       string `json:"tool_dir"`
	SoulSummary   string `json:"soul_summary"`
	MemorySummary string `json:"memory_summary"`
	Guidance      string `json:"guidance"`
}

func NewSelfAssessTool(dataDir, projectID, agentID, workspaceDir string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"self_assess",
		"Inspect your current soul, memory, dynamic tool directory, and recent self-improvement state. Use this when you notice a capability gap, repeated failure, or prompt/tool mismatch before patching your prompt or creating a new tool.",
		func(ctx context.Context, input SelfAssessInput) (*SelfAssessOutput, error) {
			agentDir := filepath.Join(dataDir, "agents", agentID)
			soulPath := filepath.Join(agentDir, "soul.md")
			memoryPath := filepath.Join(agentDir, "memory.md")
			toolDir := filepath.Join(workspaceDir, ".athena", "tools")
			soulSummary := readFileSummary(soulPath, 2000)
			memorySummary := readFileSummary(memoryPath, 2000)
			guidance := "If the gap is stable role behavior, call prompt_patch. If it is reusable executable capability, call tool_create_python. If it is one-off analysis, call python. Record evidence to blackboard after acting."
			if strings.TrimSpace(input.Focus) != "" {
				guidance = "Focus: " + input.Focus + "\n" + guidance
			}
			return &SelfAssessOutput{SoulPath: soulPath, MemoryPath: memoryPath, ToolDir: toolDir, SoulSummary: soulSummary, MemorySummary: memorySummary, Guidance: guidance}, nil
		},
	)
}

type PromptPatchInput struct {
	Title     string `json:"title" jsonschema:"description=Short title for this prompt improvement,required"`
	Rationale string `json:"rationale" jsonschema:"description=Evidence-backed reason: what failed or what capability gap was found,required"`
	Patch     string `json:"patch" jsonschema:"description=Prompt text to append to this agent's soul.md. Use durable behavior rules, not temporary task progress.,required"`
}

type PromptPatchOutput struct {
	SoulPath string `json:"soul_path"`
	Message  string `json:"message"`
}

func NewPromptPatchTool(dataDir, agentID string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"prompt_patch",
		"Append an evidence-backed prompt improvement to this agent's soul.md. Use only for durable self-improvement after self_assess; do not store temporary task progress here.",
		func(ctx context.Context, input PromptPatchInput) (*PromptPatchOutput, error) {
			if strings.TrimSpace(input.Title) == "" || strings.TrimSpace(input.Rationale) == "" || strings.TrimSpace(input.Patch) == "" {
				return nil, fmt.Errorf("title, rationale, and patch are required")
			}
			agentDir := filepath.Join(dataDir, "agents", agentID)
			if err := os.MkdirAll(agentDir, 0755); err != nil {
				return nil, fmt.Errorf("create agent dir: %w", err)
			}
			soulPath := filepath.Join(agentDir, "soul.md")
			entry := fmt.Sprintf("\n\n# 自我改进：%s\n\n## 依据\n%s\n\n## 新增规则\n%s\n", strings.TrimSpace(input.Title), strings.TrimSpace(input.Rationale), strings.TrimSpace(input.Patch))
			f, err := os.OpenFile(soulPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				return nil, fmt.Errorf("open soul: %w", err)
			}
			defer f.Close()
			if _, err := f.WriteString(entry); err != nil {
				return nil, fmt.Errorf("write soul patch: %w", err)
			}
			return &PromptPatchOutput{SoulPath: soulPath, Message: "prompt patch appended; future prompt refresh will include it"}, nil
		},
	)
}

func readFileSummary(path string, maxLen int) string {
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return ""
		}
		return fmt.Sprintf("read error: %v", err)
	}
	return truncateOutput(string(data), maxLen)
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

// --- Submit For Review Tool ---

// SubmitForReviewInput is the input for the submit_for_review tool
type SubmitForReviewInput struct {
	TaskID  string `json:"task_id" jsonschema:"description=The task ID that was assigned to you,required"`
	Summary string `json:"summary" jsonschema:"description=Brief summary of what was completed,required"`
	Files   string `json:"files" jsonschema:"description=Comma-separated list of files you produced (relative paths),required"`
}

// SubmitForReviewOutput is the output for the submit_for_review tool
type SubmitForReviewOutput struct {
	Round   int    `json:"round"`
	Message string `json:"message"`
}

// NewSubmitForReviewTool creates a tool for developers to submit completed work for PM review.
// This is the explicit signal that triggers PM verification — without it, PM stays idle.
func NewSubmitForReviewTool(dataDir, projectID, agentID string, notifyPM func(projectID, message string) error) (tool.InvokableTool, error) {
	return utils.InferTool(
		"submit_for_review",
		"Submit your completed work for PM review. This notifies the PM to start verification. You MUST use this after completing a task — PM will NOT check your work otherwise.",
		func(ctx context.Context, input SubmitForReviewInput) (*SubmitForReviewOutput, error) {
			// 1. Count existing verification rounds from blackboard
			board, err := blackboard.OpenBoard(dataDir, projectID)
			if err != nil {
				return nil, fmt.Errorf("open blackboard: %w", err)
			}

			entries, _ := board.ReadEntries("verification", 200, 0)
			round := len(entries) + 1

			// 2. Write review_pending entry to blackboard
			board.WriteEntrySync(&db.BlackboardEntry{
				ID:        uuid.New().String()[:8],
				ProjectID: projectID,
				Category:  "verification",
				Content:   fmt.Sprintf("[Round %d] Developer %s 提交验收 — Task: %s, Files: %s, Summary: %s", round, agentID, input.TaskID, input.Files, input.Summary),
				Certainty: "certain",
				Author:    agentID,
			})
			board.Close()

			// 3. Notify PM via callback (sends SteerCh message)
			if notifyPM != nil {
				msg := fmt.Sprintf("[验收通知] Developer %s 已提交验收 (Round %d)\nTask: %s\n产出文件: %s\n摘要: %s\n\n请按照验收流程执行：读取验收标准，读取产出文件，逐条对照验证。如不通过，使用 assign_task 发送整改要求。",
					agentID, round, input.TaskID, input.Files, input.Summary)
				if err := notifyPM(projectID, msg); err != nil {
					return &SubmitForReviewOutput{
						Round:   round,
						Message: fmt.Sprintf("已提交验收 (Round %d)，但PM通知失败: %v", round, err),
					}, nil
				}
			}

			return &SubmitForReviewOutput{
				Round:   round,
				Message: fmt.Sprintf("已提交验收 (Round %d)，PM已收到通知。等待验收结果...", round),
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
