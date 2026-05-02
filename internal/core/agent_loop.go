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

	// Callbacks for tools (injected by AgentManager)
	TaskFunc func(agentID, taskID, content, fromAgent string) error
	HireFunc func(req *hr.HireRequest) (*db.Agent, error)
	MainDB   *db.DB
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

	// Role-specific tools
	workspaceDir := filepath.Join(al.cfg.DataDir, "workspace", al.cfg.ProjectID)

	switch al.cfg.Role {
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
		// PM also gets file tools for workspace management
		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, fileRead)
		}
		fileWrite, err := tools.NewFileWriteTool(workspaceDir)
		if err == nil {
			agentTools = append(agentTools, fileWrite)
		}

	case "developer", "tester":
		// Developer and tester get term, file_read, file_write
		termTool, err := tools.NewTermExecTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create term tool: %w", err)
		}
		agentTools = append(agentTools, termTool)

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

	case "reviewer":
		// Reviewer gets file_read only (no write, no term)
		fileRead, err := tools.NewFileReadTool(workspaceDir)
		if err != nil {
			return nil, fmt.Errorf("create file_read tool: %w", err)
		}
		agentTools = append(agentTools, fileRead)

	case "designer":
		// Designer gets file_read and file_write
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

// buildSystemPrompt constructs the system prompt for this agent
func (al *AgentLoop) buildSystemPrompt() string {
	var sb strings.Builder

	// Layer 1: Agent identity
	sb.WriteString(fmt.Sprintf("你是 Athena 系统中的 %s Agent (ID: %s)。\n", al.cfg.Role, al.cfg.AgentID))
	sb.WriteString(fmt.Sprintf("你正在为项目 %s 工作。\n\n", al.cfg.ProjectID))

	// Layer 2: Role-specific instructions
	switch al.cfg.Role {
	case "developer":
		sb.WriteString("你是一名专业的软件开发工程师。你的职责是编写高质量、可维护的代码。\n")
		sb.WriteString("你只负责开发，不负责测试、设计或审查。\n")
		sb.WriteString("将你的工作进展写入黑板，将发现的事实标记为\"确定\"或\"猜测\"。\n")
		sb.WriteString("遇到别人领域的问题（如测试bug、设计疑问），立刻找对应Agent开会对齐，绝不自己琢磨。\n")
		sb.WriteString("你可以使用 term 工具执行命令，使用 file_write 工具创建和修改文件。\n")
		sb.WriteString("完成开发后，将代码文件路径和关键说明写入黑板。\n")
	case "tester":
		sb.WriteString("你是一名专业的测试工程师。你的职责是编写测试用例、执行测试、出具测试报告。\n")
		sb.WriteString("你只负责测试，不负责开发或审查。\n")
		sb.WriteString("发现bug时，立刻告诉对应项目的开发Agent，并详细传递：报错信息、环境、测试方法、测试用例。\n")
	case "pm":
		sb.WriteString("你是项目经理Agent。你的职责是拆解细化需求、分配任务、验收交付（最高标准）。\n")
		sb.WriteString("验收时必须进行\"需求回溯\"——对照CEO原始需求逐条确认。\n")
		sb.WriteString("验收不通过→要求整改→循环验证直至完善→避免CEO返工。\n")
		sb.WriteString("你需要招聘人手时，使用 hr_request 工具。\n")
		sb.WriteString("分配任务时，使用 assign_task 工具指定角色和任务内容。\n")
		sb.WriteString("工作流程：\n")
		sb.WriteString("1. 分析CEO需求，拆解为具体任务\n")
		sb.WriteString("2. 使用 hr_request 招聘需要的角色（如developer）\n")
		sb.WriteString("3. 使用 assign_task 分配任务给对应角色\n")
		sb.WriteString("4. 读取黑板跟踪进展\n")
		sb.WriteString("5. 验收交付，确保需求回溯完整\n")
	case "reviewer":
		sb.WriteString("你是代码审查员。你的职责是以最高标准审核所有代码变更。\n")
		sb.WriteString("你的上下文与开发、测试隔离，只基于原始代码和原始需求审查。\n")
		sb.WriteString("审查维度：代码正确性、健壮性、性能、安全性、可维护性、边界条件、异常处理。\n")
	default:
		sb.WriteString(fmt.Sprintf("你是角色为 %s 的专业Agent，专注本职工作。\n", al.cfg.Role))
	}

	// Layer 3: Blackboard usage instructions
	sb.WriteString("\n## 黑板使用规则\n")
	sb.WriteString("- 读取黑板获取项目目标、事实、进展\n")
	sb.WriteString("- 写入黑板分享你的工作成果和新发现\n")
	sb.WriteString("- 确定性事实标记为\"certain\"（必须100%可靠），猜测标记为\"conjecture\"\n")
	sb.WriteString("- 每次推理后进行自省评分（0-10分）\n")
	sb.WriteString("- 10分结果必须附带完整推理流程\n")

	return sb.String()
}

// BuildAgentPrompt builds the full system prompt for an agent (used by context_builder)
func BuildAgentPrompt(agentID, role, projectID string, board *blackboard.Board) string {
	loop := &AgentLoop{
		cfg: &AgentLoopConfig{
			AgentID:   agentID,
			Role:      role,
			ProjectID: projectID,
		},
	}

	prompt := loop.buildSystemPrompt()

	// Layer 3: Project-level context from blackboard
	if board != nil {
		goals, _ := board.ReadEntries(blackboard.CategoryGoal, 20, 0)
		facts, _ := board.ReadEntries(blackboard.CategoryFact, 20, 0)
		resolutions, _ := board.ReadEntries(blackboard.CategoryResolution, 10, 0)

		if len(goals) > 0 {
			prompt += "\n## 项目目标\n"
			for _, g := range goals {
				prompt += fmt.Sprintf("- [%s] %s\n", g.Certainty, g.Content)
			}
		}

		if len(facts) > 0 {
			prompt += "\n## 确定性事实\n"
			for _, f := range facts {
				prompt += fmt.Sprintf("- [%s] %s\n", f.Certainty, f.Content)
			}
		}

		if len(resolutions) > 0 {
			prompt += "\n## 会议决议\n"
			for _, r := range resolutions {
				prompt += fmt.Sprintf("- %s\n", r.Content)
			}
		}
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
