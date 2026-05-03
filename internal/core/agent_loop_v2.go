package core

import (
	"context"
	"fmt"
	"strings"

	"github.com/cloudwego/eino/components/tool"
	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/db"
)

// RunInProcess runs the agent loop inside a goroutine (no subprocess)
// It reads from the AgentHandle channels instead of stdin/stdout
func (al *AgentLoop) RunInProcess(ctx context.Context, handle *AgentHandle) error {
	// 1. Use shared LLMClient — get primary ChatModel for Eino tool binding
	chatModel := al.cfg.LLM.PrimaryChatModel()

	// 2. Create Athena tools
	agentTools, err := al.createTools(ctx)
	if err != nil {
		return fmt.Errorf("create tools: %w", err)
	}

	// 3. Build system prompt (6-layer architecture + blackboard context)
	systemPrompt := al.buildSystemPrompt()

	// 4. Get tool infos
	toolInfos := al.getToolInfos(ctx, agentTools)

	// Conversation history
	messages := []*schema.Message{schema.SystemMessage(systemPrompt)}

	for {
		select {
		case <-ctx.Done():
			al.logger.Println("context cancelled, agent exiting")
			return nil

		case task := <-handle.TaskCh:
			al.logger.Printf("received task: %s", truncateStr(task.Content, 80))
			handle.Status = StatusWorking

			// Build user message
			userMsg := fmt.Sprintf("任务: %s\n\n请按照你的工作流程执行。使用工具完成工作，将结果写入黑板。", task.Content)
			messages = append(messages, schema.UserMessage(userMsg))

			// ReAct loop
			al.runReActLoop(ctx, &messages, chatModel, agentTools, toolInfos, 20)

			handle.Status = StatusIdle

		case steer := <-handle.SteerCh:
			al.logger.Printf("received steer: %s", truncateStr(steer, 80))
			handle.Status = StatusWorking

			// Differentiate steer type
			var userMsg string
			if strings.HasPrefix(steer, "[验收通知]") {
				// PM verification notification — inject verification context
				userMsg = al.buildVerificationSteerMessage(steer)
			} else {
				// CEO new requirement
				userMsg = fmt.Sprintf("[CEO新需求] %s\n\n请评估此需求对当前工作的影响，更新黑板。", steer)
			}
			messages = append(messages, schema.UserMessage(userMsg))

			// ReAct loop for steer response
			al.runReActLoop(ctx, &messages, chatModel, agentTools, toolInfos, 20)

			handle.Status = StatusIdle
		}
	}
}

// runReActLoop executes the ReAct reasoning-action cycle
func (al *AgentLoop) runReActLoop(
	ctx context.Context,
	messages *[]*schema.Message,
	chatModel einomodel.ChatModel,
	agentTools []tool.InvokableTool,
	toolInfos []*schema.ToolInfo,
	maxIterations int,
) {
	for i := 0; i < maxIterations; i++ {
		var opts []einomodel.Option
		if len(toolInfos) > 0 {
			opts = append(opts, einomodel.WithTools(toolInfos))
		}

		response, err := chatModel.Generate(ctx, *messages, opts...)
		if err != nil {
			al.logger.Printf("LLM error: %v", err)
			break
		}

		*messages = append(*messages, response)

		// No tool calls — agent is done with this iteration
		if len(response.ToolCalls) == 0 {
			// Write the final result to the blackboard
			board, boardErr := blackboard.OpenBoard(al.cfg.DataDir, al.cfg.ProjectID)
			if boardErr == nil {
				board.WriteEntrySync(&db.BlackboardEntry{
					ID:        generateUUID(),
					ProjectID: al.cfg.ProjectID,
					Category:  "progress",
					Content:   fmt.Sprintf("[Agent %s 完成当前轮次] %s", al.cfg.AgentID, truncateStr(cleanContent(response.Content), 500)),
					Certainty: "certain",
					Author:    al.cfg.AgentID,
				})
				board.Close()
			}
			break
		}

		// Process tool calls
		for _, tc := range response.ToolCalls {
			result, err := al.executeToolCall(ctx, agentTools, tc)
			if err != nil {
				result = fmt.Sprintf("Tool error: %v", err)
			}
			*messages = append(*messages, schema.ToolMessage(result, tc.ID))
		}
	}
}

// buildVerificationSteerMessage constructs a detailed steer message for PM verification
func (al *AgentLoop) buildVerificationSteerMessage(steer string) string {
	var sb strings.Builder
	sb.WriteString(steer)
	sb.WriteString("\n\n---\n\n")

	// Inject current verification round count
	board, err := blackboard.OpenBoard(al.cfg.DataDir, al.cfg.ProjectID)
	if err == nil {
		verifications, _ := board.ReadEntries(blackboard.CategoryVerification, 200, 0)
		round := len(verifications)

		sb.WriteString("## 当前验收状态\n")
		sb.WriteString(fmt.Sprintf("- 已完成验收轮次: %d / 100\n", round))

		if round >= 100 {
			sb.WriteString("\n⚠️ **验收已达100轮上限！** 必须执行以下操作：\n")
			sb.WriteString("1. 使用 blackboard_write 写入验收超限上报（category: \"verification\"，content 包含 \"[ESCALATION]\"）\n")
			sb.WriteString("2. 内容需包含：累计问题清单、各轮主要问题摘要、建议CEO决策方向\n")
			sb.WriteString("3. 写完上报后停止验收循环，等待CEO决策\n")
		} else if round >= 80 {
			sb.WriteString(fmt.Sprintf("\n⚠️ 验收已进行 %d 轮，接近100轮上限。请评估是否需要上报CEO。\n", round))
		} else {
			sb.WriteString("\n请按照验收流程执行：\n")
			sb.WriteString("1. 使用 blackboard_read 读取验收标准（category: \"acceptance_criteria\"）\n")
			sb.WriteString("2. 使用 file_read 读取 developer 的产出文件\n")
			sb.WriteString("3. 逐条对照验收标准，记录通过/不通过\n")
			sb.WriteString("4. 如全部通过 → 使用 blackboard_write 写入验收通过报告（category: \"verification\"，content 包含 \"[PASS]\"）\n")
			sb.WriteString("5. 如有不通过项 → 使用 assign_task 发送整改要求（附具体问题清单）\n")
		}

		board.Close()
	}

	return sb.String()
}
