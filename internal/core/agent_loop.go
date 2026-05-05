package core

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/db"
	"github.com/ksroido/athena/internal/hr"
	"github.com/ksroido/athena/internal/tools"
)

// AgentLoopConfig holds configuration for an agent's ReAct loop
type AgentLoopConfig struct {
	AgentID   string
	Role      string
	ProjectID string
	DataDir   string
	LLM       *LLMClient // Reuse the shared LLMClient
	HR        *hr.HR     // HR instance for role resolution

	// Callbacks for tools (injected by AgentManager)
	TaskFunc     func(agentID, taskID, content, fromAgent string) error
	HireFunc     func(req *hr.HireRequest) (*db.Agent, error)
	NotifyPMFunc func(projectID, message string) error
	MainDB       *db.DB
}

// AgentLoop runs a single Agent's ReAct loop using Eino ChatModelAgent
type AgentLoop struct {
	cfg    *AgentLoopConfig
	logger *log.Logger
}

// NewAgentLoop creates a new AgentLoop
func NewAgentLoop(cfg *AgentLoopConfig) *AgentLoop {
	return &AgentLoop{
		cfg:    cfg,
		logger: log.New(os.Stderr, fmt.Sprintf("[agent:%s] ", cfg.AgentID), log.LstdFlags),
	}
}

// createTools creates the Eino tools for this agent
func (al *AgentLoop) createTools(ctx context.Context) ([]tool.InvokableTool, error) {
	var agentTools []tool.InvokableTool
	workspaceDir := filepath.Join(al.cfg.DataDir, "workspace", al.cfg.ProjectID)

	// Blackboard read tool
	bbRead, err := tools.NewBlackboardReadTool(al.cfg.DataDir, al.cfg.ProjectID, al.cfg.Role)
	if err != nil {
		return nil, fmt.Errorf("create blackboard read tool: %w", err)
	}
	agentTools = append(agentTools, bbRead)

	// Blackboard write tool
	bbWrite, err := tools.NewBlackboardWriteTool(al.cfg.DataDir, al.cfg.ProjectID, al.cfg.AgentID, al.cfg.Role)
	if err != nil {
		return nil, fmt.Errorf("create blackboard write tool: %w", err)
	}
	agentTools = append(agentTools, bbWrite)

	// Memory tools (file-based)
	memRead, err := tools.NewMemoryReadToolFile(al.cfg.DataDir, al.cfg.AgentID)
	if err != nil {
		return nil, fmt.Errorf("create memory read tool: %w", err)
	}
	agentTools = append(agentTools, memRead)

	memWrite, err := tools.NewMemoryWriteToolFile(al.cfg.DataDir, al.cfg.AgentID)
	if err != nil {
		return nil, fmt.Errorf("create memory write tool: %w", err)
	}
	agentTools = append(agentTools, memWrite)

	// Meeting tool
	meeting, err := tools.NewMeetingTool(al.cfg.AgentID, al.cfg.ProjectID)
	if err != nil {
		return nil, fmt.Errorf("create meeting tool: %w", err)
	}
	agentTools = append(agentTools, meeting)

	// Self-improvement tools are available to all roles. They let an agent record
	// durable prompt gaps and create reusable Python capabilities instead of
	// relying on hardcoded role-specific solutions.
	selfAssess, err := tools.NewSelfAssessTool(al.cfg.DataDir, al.cfg.ProjectID, al.cfg.AgentID, workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("create self_assess tool: %w", err)
	}
	agentTools = append(agentTools, selfAssess)

	promptPatch, err := tools.NewPromptPatchTool(al.cfg.DataDir, al.cfg.AgentID)
	if err != nil {
		return nil, fmt.Errorf("create prompt_patch tool: %w", err)
	}
	agentTools = append(agentTools, promptPatch)

	dynamicRunner, err := tools.NewDynamicPythonToolRunner(workspaceDir)
	if err != nil {
		return nil, fmt.Errorf("create dynamic_python_tool: %w", err)
	}
	agentTools = append(agentTools, dynamicRunner)

	// Role-specific tools
	category := hr.InferCategory(al.cfg.Role)

	switch category {
	case "pm":
		// PM gets assign_task and hr_request
		if al.cfg.TaskFunc != nil && al.cfg.MainDB != nil {
			assignTask, err := tools.NewAssignTaskTool(al.cfg.ProjectID, al.cfg.AgentID, al.cfg.MainDB, al.cfg.TaskFunc, al.cfg.HireFunc)
			if err != nil {
				return nil, fmt.Errorf("create assign_task tool: %w", err)
			}
			agentTools = append(agentTools, assignTask)
		}
		if al.cfg.HireFunc != nil {
			hrReq, err := tools.NewHRRequestTool(al.cfg.ProjectID, al.cfg.HireFunc)
			if err != nil {
				return nil, fmt.Errorf("create hr_request tool: %w", err)
			}
			agentTools = append(agentTools, hrReq)
		}
		// PM also gets file tools for verification (must read developer output)
		pythonTool, err := tools.NewPythonExecTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, pythonTool)
		}
		toolCreate, err := tools.NewPythonToolCreateTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, toolCreate)
		}
		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, fileRead)
		}
		fileWrite, err := tools.NewFileWriteTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, fileWrite)
		}

	case "dev":
		// All dev.* roles get term, python, dynamic tool authoring, file_read, file_write, submit_for_review
		termTool, err := tools.NewTermExecTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create term tool: %w", err)
		}
		agentTools = append(agentTools, termTool)

		pythonTool, err := tools.NewPythonExecTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create python tool: %w", err)
		}
		agentTools = append(agentTools, pythonTool)

		toolCreate, err := tools.NewPythonToolCreateTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create tool_create_python: %w", err)
		}
		agentTools = append(agentTools, toolCreate)

		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create file_read tool: %w", err)
		}
		agentTools = append(agentTools, fileRead)

		fileWrite, err := tools.NewFileWriteTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create file_write tool: %w", err)
		}
		agentTools = append(agentTools, fileWrite)

		// Dev roles get submit_for_review to trigger PM verification
		if al.cfg.NotifyPMFunc != nil {
			submitReview, err := tools.NewSubmitForReviewTool(al.cfg.DataDir, al.cfg.ProjectID, al.cfg.AgentID, al.cfg.NotifyPMFunc)
			if err != nil {
				return nil, fmt.Errorf("create submit_for_review tool: %w", err)
			}
			agentTools = append(agentTools, submitReview)
		}

	case "tester":
		// All tester* roles get term, python, dynamic tool authoring, file_read, file_write
		termTool, err := tools.NewTermExecTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create term tool: %w", err)
		}
		agentTools = append(agentTools, termTool)

		pythonTool, err := tools.NewPythonExecTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create python tool: %w", err)
		}
		agentTools = append(agentTools, pythonTool)

		toolCreate, err := tools.NewPythonToolCreateTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create tool_create_python: %w", err)
		}
		agentTools = append(agentTools, toolCreate)

		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create file_read tool: %w", err)
		}
		agentTools = append(agentTools, fileRead)

		fileWrite, err := tools.NewFileWriteTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create file_write tool: %w", err)
		}
		agentTools = append(agentTools, fileWrite)

		// Tester roles also get submit_for_review
		if al.cfg.NotifyPMFunc != nil {
			submitReview, err := tools.NewSubmitForReviewTool(al.cfg.DataDir, al.cfg.ProjectID, al.cfg.AgentID, al.cfg.NotifyPMFunc)
			if err != nil {
				return nil, fmt.Errorf("create submit_for_review tool: %w", err)
			}
			agentTools = append(agentTools, submitReview)
		}

	case "reviewer":
		// Reviewer gets file_read only (no write, no term)
		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create file_read tool: %w", err)
		}
		agentTools = append(agentTools, fileRead)

	case "designer":
		// All designer* roles get file_read and file_write
		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, fileRead)
		}
		fileWrite, err := tools.NewFileWriteTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, fileWrite)
		}
	}

	return agentTools, nil
}

// getToolInfos extracts ToolInfo from all registered tools
func (al *AgentLoop) getToolInfos(ctx context.Context, agentTools []tool.InvokableTool) []*schema.ToolInfo {
	var infos []*schema.ToolInfo
	for _, t := range agentTools {
		info, err := t.Info(ctx)
		if err == nil && info != nil {
			infos = append(infos, info)
		}
	}
	return infos
}

// executeToolCall finds and executes the appropriate tool
func (al *AgentLoop) executeToolCall(ctx context.Context, agentTools []tool.InvokableTool, tc schema.ToolCall) (string, error) {
	for _, t := range agentTools {
		info, err := t.Info(ctx)
		if err != nil {
			continue
		}
		if info.Name == tc.Function.Name {
			return t.InvokableRun(ctx, tc.Function.Arguments)
		}
	}
	return "", fmt.Errorf("tool %s not found", tc.Function.Name)
}

// buildAvailableToolsSection formats the actual tools bound to this agent.
func (al *AgentLoop) buildAvailableToolsSection(ctx context.Context, toolInfos []*schema.ToolInfo) string {
	var sb strings.Builder
	sb.WriteString("\n# 当前可用工具\n\n")
	sb.WriteString("只能调用本节列出的工具。不要假设存在未列出的工具。工具返回值是事实依据；失败时必须读取错误信息并记录处理策略。\n")
	sb.WriteString("如果发现工具不足，先使用 self_assess 明确能力缺口；稳定行为缺口用 prompt_patch 修订 soul；可复用执行能力用 tool_create_python 创建动态 Python 工具；一次性计算或探查用 python。\n\n")
	for _, info := range toolInfos {
		desc := strings.TrimSpace(info.Desc)
		if desc == "" {
			desc = "无描述"
		}
		sb.WriteString(fmt.Sprintf("- `%s`: %s\n", info.Name, desc))
	}
	workspaceDir := filepath.Join(al.cfg.DataDir, "workspace", al.cfg.ProjectID)
	if dynamicTools := tools.DynamicPythonToolInventory(workspaceDir); strings.TrimSpace(dynamicTools) != "" {
		sb.WriteString("\n## 已注册动态 Python 工具\n")
		sb.WriteString("通过 `dynamic_python_tool` 按 name 调用：\n")
		sb.WriteString(dynamicTools)
	}
	sb.WriteString("\n")
	return sb.String()
}

func (al *AgentLoop) refreshToolsForTurn(ctx context.Context) ([]tool.InvokableTool, []*schema.ToolInfo) {
	agentTools, err := al.createTools(ctx)
	if err != nil {
		al.logger.Printf("refresh tools error: %v", err)
		return nil, nil
	}
	return agentTools, al.getToolInfos(ctx, agentTools)
}

func refreshSystemMessage(messages []*schema.Message, systemPrompt string) []*schema.Message {
	msg := schema.SystemMessage(systemPrompt)
	if len(messages) == 0 {
		return []*schema.Message{msg}
	}
	messages[0] = msg
	return messages
}

// buildSystemPrompt constructs the system prompt using the role prompt architecture
func (al *AgentLoop) buildSystemPrompt() string {
	_, toolInfos := al.refreshToolsForTurn(context.Background())
	return al.buildSystemPromptWithTools(context.Background(), toolInfos)
}

func (al *AgentLoop) buildSystemPromptWithTools(ctx context.Context, toolInfos []*schema.ToolInfo) string {
	// Role prompt from prompts.go (with dynamic soul resolution)
	prompt := BuildRolePrompt(al.cfg.Role, al.cfg.AgentID, al.cfg.ProjectID, al.cfg.DataDir, al.cfg.HR)

	// Append Layer 7: actual tool inventory bound to this agent.
	prompt += al.buildAvailableToolsSection(ctx, toolInfos)

	// Append Layer 8: Project context from blackboard
	prompt += buildBlackboardContext(al.cfg.DataDir, al.cfg.ProjectID)

	// Append Layer 9: Available roles catalog (for PM)
	if al.cfg.Role == "pm" && al.cfg.HR != nil {
		prompt += "\n# 可招聘角色\n\n"
		prompt += al.cfg.HR.RoleCatalog(al.cfg.ProjectID)
	}

	return prompt
}

// buildBlackboardContext reads the blackboard and formats project context
func buildBlackboardContext(dataDir, projectID string) string {
	board, err := blackboard.OpenBoard(dataDir, projectID)
	if err != nil {
		return ""
	}
	defer board.Close()

	var sb strings.Builder
	sb.WriteString("\n# 项目上下文\n\n")

	// Goals
	goals, _ := board.ReadEntries(blackboard.CategoryGoal, 20, 0)
	if len(goals) > 0 {
		sb.WriteString("## 项目目标\n")
		for _, g := range goals {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", g.Certainty, g.Content))
		}
		sb.WriteString("\n")
	}

	// Acceptance criteria (critical for PM verification)
	criteria, _ := board.ReadEntries(blackboard.CategoryAcceptanceCrit, 30, 0)
	if len(criteria) > 0 {
		sb.WriteString("## 验收标准\n")
		for _, c := range criteria {
			sb.WriteString(fmt.Sprintf("- %s\n", c.Content))
		}
		sb.WriteString("\n")
	}

	// Facts
	facts, _ := board.ReadEntries(blackboard.CategoryFact, 20, 0)
	if len(facts) > 0 {
		sb.WriteString("## 确定性事实\n")
		for _, f := range facts {
			sb.WriteString(fmt.Sprintf("- [%s] %s\n", f.Certainty, f.Content))
		}
		sb.WriteString("\n")
	}

	// Verification history (for PM to know current round)
	verifications, _ := board.ReadEntries(blackboard.CategoryVerification, 200, 0)
	if len(verifications) > 0 {
		sb.WriteString(fmt.Sprintf("## 验收记录（共 %d 轮）\n", len(verifications)))
		// Show last 5 rounds to avoid context bloat
		start := 0
		if len(verifications) > 5 {
			start = len(verifications) - 5
			sb.WriteString(fmt.Sprintf("（仅显示最近5轮，总计%d轮）\n", len(verifications)))
		}
		for i := start; i < len(verifications); i++ {
			sb.WriteString(fmt.Sprintf("- %s\n", verifications[i].Content))
		}
		sb.WriteString("\n")
	}

	// Resolutions
	resolutions, _ := board.ReadEntries(blackboard.CategoryResolution, 10, 0)
	if len(resolutions) > 0 {
		sb.WriteString("## 会议决议\n")
		for _, r := range resolutions {
			sb.WriteString(fmt.Sprintf("- %s\n", r.Content))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// BuildAgentPrompt builds the full system prompt for an agent (used by context_builder)
func BuildAgentPrompt(agentID, role, projectID, dataDir string, hrInstance *hr.HR, board *blackboard.Board) string {
	prompt := BuildRolePrompt(role, agentID, projectID, dataDir, hrInstance)

	// Append blackboard context
	if board != nil {
		var sb strings.Builder
		sb.WriteString("\n# 项目上下文\n\n")

		goals, _ := board.ReadEntries(blackboard.CategoryGoal, 20, 0)
		if len(goals) > 0 {
			sb.WriteString("## 项目目标\n")
			for _, g := range goals {
				sb.WriteString(fmt.Sprintf("- [%s] %s\n", g.Certainty, g.Content))
			}
			sb.WriteString("\n")
		}

		criteria, _ := board.ReadEntries(blackboard.CategoryAcceptanceCrit, 30, 0)
		if len(criteria) > 0 {
			sb.WriteString("## 验收标准\n")
			for _, c := range criteria {
				sb.WriteString(fmt.Sprintf("- %s\n", c.Content))
			}
			sb.WriteString("\n")
		}

		facts, _ := board.ReadEntries(blackboard.CategoryFact, 20, 0)
		if len(facts) > 0 {
			sb.WriteString("## 确定性事实\n")
			for _, f := range facts {
				sb.WriteString(fmt.Sprintf("- [%s] %s\n", f.Certainty, f.Content))
			}
			sb.WriteString("\n")
		}

		resolutions, _ := board.ReadEntries(blackboard.CategoryResolution, 10, 0)
		if len(resolutions) > 0 {
			sb.WriteString("## 会议决议\n")
			for _, r := range resolutions {
				sb.WriteString(fmt.Sprintf("- %s\n", r.Content))
			}
		}

		prompt += sb.String()
	}

	return prompt
}

// truncateStr truncates a string for logging
func truncateStr(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen] + "..."
}

// cleanContent removes LLM-specific tags (thinking, tool_call markers) from
// response content before writing to the blackboard or returning to users.
// Many models (DeepSeek, GLM, Qwen) emit <think>...</think> blocks or
// similar markers in their text output that should not leak into shared state.
func cleanContent(s string) string {
	// Remove <think>...</think> blocks (including multiline)
	for {
		start := strings.Index(s, "<think>")
		if start == -1 {
			break
		}
		end := strings.Index(s, "</think>")
		if end == -1 || end < start {
			// Unclosed tag — remove from <think> to end
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("</think>"):]
	}

	// Remove <thinking>...</thinking> blocks
	for {
		start := strings.Index(s, "<thinking>")
		if start == -1 {
			break
		}
		end := strings.Index(s, "</thinking>")
		if end == -1 || end < start {
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("</thinking>"):]
	}

	// Remove ◁think▷...◁/think▷ blocks (some models use non-ASCII markers)
	for {
		start := strings.Index(s, "◁think▷")
		if start == -1 {
			break
		}
		end := strings.Index(s, "◁/think▷")
		if end == -1 || end < start {
			s = s[:start]
			break
		}
		s = s[:start] + s[end+len("◁/think▷"):]
	}

	// Clean up excessive whitespace left after tag removal
	s = strings.TrimSpace(s)

	return s
}
