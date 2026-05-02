package core

import (
	"context"
	"fmt"

	einomodel "github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/db"
)

// RunInProcess runs the agent loop inside a goroutine (no subprocess)
// It reads from the AgentHandle channels instead of stdin/stdout
func (al *AgentLoop) RunInProcess(ctx context.Context, handle *AgentHandle) error {
	// 1. Use shared LLMClient
	chatModel := al.cfg.LLM.chatModel

	// 2. Create Athena tools
	agentTools, err := al.createTools(ctx)
	if err != nil {
		return fmt.Errorf("create tools: %w", err)
	}

	// 3. Build system prompt
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
			userMsg := fmt.Sprintf("任务: %s\n\n请分析任务并开始执行。使用你的工具完成工作，将结果写入黑板。", task.Content)
			messages = append(messages, schema.UserMessage(userMsg))

			// ReAct loop
			maxIterations := 15
			for i := 0; i < maxIterations; i++ {
				var opts []einomodel.Option
				if len(toolInfos) > 0 {
					opts = append(opts, einomodel.WithTools(toolInfos))
				}

				response, err := chatModel.Generate(ctx, messages, opts...)
				if err != nil {
					al.logger.Printf("LLM error: %v", err)
					break
				}

				messages = append(messages, response)

				// No tool calls — agent is done
				if len(response.ToolCalls) == 0 {
					// Write the final result to the blackboard
					board, boardErr := blackboard.OpenBoard(al.cfg.DataDir, al.cfg.ProjectID)
					if boardErr == nil {
						board.WriteEntrySync(&db.BlackboardEntry{
							ID:        generateUUID(),
							ProjectID: al.cfg.ProjectID,
							Category:  "progress",
							Content:   fmt.Sprintf("[Agent %s 完成任务] %s", al.cfg.AgentID, response.Content),
							Certainty: "certain",
							Author:    al.cfg.AgentID,
						})
						board.Close()
					}
					handle.Status = StatusIdle
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

			handle.Status = StatusIdle

		case steer := <-handle.SteerCh:
			al.logger.Printf("received steer: %s", truncateStr(steer, 80))
			messages = append(messages, schema.UserMessage(
				fmt.Sprintf("[CEO新需求] %s\n\n请评估此需求对当前工作的影响，更新黑板。", steer),
			))
		}
	}
}
