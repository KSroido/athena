package core

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"strings"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/tool"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/tools"
)

// AgentLoopConfig holds configuration for an agent's ReAct loop
type AgentLoopConfig struct {
	AgentID    string
	Role       string
	ProjectID  string
	DataDir    string
	LLMBaseURL string
	LLMAPIKey  string
	LLMModel   string
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

// Run starts the agent loop, reading from stdin and writing to stdout
// This implements the athena-agent subprocess protocol
func (al *AgentLoop) Run(ctx context.Context, in io.Reader, out io.Writer) error {
	// 1. Create Eino ChatModel
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: al.cfg.LLMBaseURL,
		APIKey:  al.cfg.LLMAPIKey,
		Model:   al.cfg.LLMModel,
	})
	if err != nil {
		return fmt.Errorf("create chat model: %w", err)
	}

	// 2. Create Athena tools (bridged to Eino Tool interface)
	agentTools, err := al.createTools(ctx)
	if err != nil {
		return fmt.Errorf("create tools: %w", err)
	}

	// 3. Build system prompt
	systemPrompt := al.buildSystemPrompt()

	// 4. Get tool infos for ChatModel
	toolInfos := al.getToolInfos(ctx, agentTools)

	// 5. Main loop: read stdin → build messages → call LLM → write stdout
	decoder := json.NewDecoder(in)
	encoder := json.NewEncoder(out)

	// Conversation history (starts with system prompt)
	messages := []*schema.Message{schema.SystemMessage(systemPrompt)}

	for {
		var msg AgentMessage
		if err := decoder.Decode(&msg); err != nil {
			if err == io.EOF {
				al.logger.Println("stdin closed, agent exiting")
				return nil
			}
			al.logger.Printf("decode error: %v", err)
			continue
		}

		switch msg.Type {
		case "task":
			al.logger.Printf("received task: %s", truncateStr(msg.Content, 80))

			// Build user message
			userMsg := fmt.Sprintf("任务: %s\n\n请分析任务并开始执行。使用你的工具完成工作，将结果写入黑板。", msg.Content)
			messages = append(messages, schema.UserMessage(userMsg))

			// ReAct loop: call LLM → handle tool calls → call LLM again
			maxIterations := 10
			for i := 0; i < maxIterations; i++ {
				// Call LLM with tools
				var opts []einomodel.Option
				if len(toolInfos) > 0 {
					opts = append(opts, einomodel.WithTools(toolInfos))
				}

				response, err := chatModel.Generate(ctx, messages, opts...)
				if err != nil {
					al.logger.Printf("LLM error: %v", err)
					encoder.Encode(AgentResponse{
						Type:   "error",
						TaskID: msg.TaskID,
						Error:  err.Error(),
					})
					break
				}

				// Add assistant response to history
				messages = append(messages, response)

				// If no tool calls, we're done
				if len(response.ToolCalls) == 0 {
					encoder.Encode(AgentResponse{
						Type:    "task_result",
						TaskID:  msg.TaskID,
						Content: response.Content,
					})
					break
				}

				// Process tool calls
				for _, tc := range response.ToolCalls {
					result, err := al.executeToolCall(ctx, agentTools, tc)
					if err != nil {
						result = fmt.Sprintf("Tool error: %v", err)
					}
					messages = append(messages, schema.ToolMessage(result, tc.ID))
				}
			}

		case "steer":
			al.logger.Printf("received steer: %s", truncateStr(msg.Content, 80))
			messages = append(messages, schema.UserMessage(
				fmt.Sprintf("[CEO新需求] %s\n\n请评估此需求对当前工作的影响，更新黑板。", msg.Content),
			))

		case "meeting_invite":
			al.logger.Printf("received meeting invite: %s", msg.MeetingID)
			// TODO: Phase 3

		default:
			al.logger.Printf("unknown message type: %s", msg.Type)
		}
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

	// Memory tools
	memRead, err := tools.NewMemoryReadTool(al.cfg.DataDir, al.cfg.AgentID)
	if err != nil {
		return nil, fmt.Errorf("create memory read tool: %w", err)
	}
	agentTools = append(agentTools, memRead)

	memWrite, err := tools.NewMemoryWriteTool(al.cfg.DataDir, al.cfg.AgentID)
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
	case "tester":
		sb.WriteString("你是一名专业的测试工程师。你的职责是编写测试用例、执行测试、出具测试报告。\n")
		sb.WriteString("你只负责测试，不负责开发或审查。\n")
		sb.WriteString("发现bug时，立刻告诉对应项目的开发Agent，并详细传递：报错信息、环境、测试方法、测试用例。\n")
	case "pm":
		sb.WriteString("你是项目经理Agent。你的职责是拆解细化需求、分配任务、验收交付（最高标准）。\n")
		sb.WriteString("验收时必须进行\"需求回溯\"——对照CEO原始需求逐条确认。\n")
		sb.WriteString("验收不通过→要求整改→循环验证直至完善→避免CEO返工。\n")
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
