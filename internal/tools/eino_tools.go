package tools

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/google/uuid"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/db"
)

// --- Blackboard Read Tool ---

// BlackboardReadInput is the input for the blackboard read tool
type BlackboardReadInput struct {
	Category string `json:"category" jsonschema:"description=Filter by category (goal/fact/discovery/decision/progress/resolution/auxiliary). Empty for all."`
	Limit    int    `json:"limit" jsonschema:"description=Maximum entries to return. Default 50."`
	Query    string `json:"query" jsonschema:"description=FTS5 full-text search query. If provided, category filter is ignored."`
}

// BlackboardReadOutput is the output for the blackboard read tool
type BlackboardReadOutput struct {
	Entries []*db.BlackboardEntry `json:"entries"`
	Count   int                   `json:"count"`
}

// NewBlackboardReadTool creates an Eino tool for reading from the blackboard
func NewBlackboardReadTool(dataDir, projectID string, role string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"blackboard_read",
		"Read entries from the project blackboard. Supports category filtering and FTS5 full-text search.",
		func(ctx context.Context, input BlackboardReadInput) (*BlackboardReadOutput, error) {
			board, err := blackboard.OpenBoard(dataDir, projectID)
			if err != nil {
				return nil, fmt.Errorf("open blackboard: %w", err)
			}
			defer board.Close()

			limit := input.Limit
			if limit <= 0 {
				limit = 50
			}

			var entries []*db.BlackboardEntry

			if input.Query != "" {
				// FTS5 search
				entries, err = board.Search(input.Query, limit)
				if err != nil {
					return nil, fmt.Errorf("search: %w", err)
				}
			} else {
				// Category-filtered read
				entries, err = board.ReadEntries(input.Category, limit, 0)
				if err != nil {
					return nil, fmt.Errorf("read entries: %w", err)
				}
			}

			// Filter by access control
			var filtered []*db.BlackboardEntry
			for _, e := range entries {
				level := blackboard.CategoryToLevel(e.Category)
				if blackboard.CanRead(role, level) {
					filtered = append(filtered, e)
				}
			}

			return &BlackboardReadOutput{
				Entries: filtered,
				Count:   len(filtered),
			}, nil
		},
	)
}

// --- Blackboard Write Tool ---

// BlackboardWriteInput is the input for the blackboard write tool
type BlackboardWriteInput struct {
	Category        string `json:"category" jsonschema:"description=Entry category (fact/discovery/progress/decision/auxiliary),required"`
	Content         string `json:"content" jsonschema:"description=Entry content,required"`
	Certainty       string `json:"certainty" jsonschema:"description=Certainty level: certain/conjecture/pending_verification. Default: conjecture"`
	ConfidenceScore *int   `json:"confidence_score" jsonschema:"description=Self-assessment score 0-10. 10 requires reasoning."`
	Reasoning       string `json:"reasoning" jsonschema:"description=Reasoning chain for confidence_score=10 entries"`
}

// BlackboardWriteOutput is the output for the blackboard write tool
type BlackboardWriteOutput struct {
	EntryID  string `json:"entry_id"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

// NewBlackboardWriteTool creates an Eino tool for writing to the blackboard
func NewBlackboardWriteTool(dataDir, projectID, agentID, role string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"blackboard_write",
		"Write an entry to the project blackboard. Only write entries within your role's permissions. Mark certainty carefully: 'certain' only for 100% verified facts.",
		func(ctx context.Context, input BlackboardWriteInput) (*BlackboardWriteOutput, error) {
			// Check write permission
			level := blackboard.CategoryToLevel(input.Category)
			if !blackboard.CanWrite(role, level) {
				return nil, fmt.Errorf("role %s cannot write to category %s (level %d)", role, input.Category, level)
			}

			board, err := blackboard.OpenBoard(dataDir, projectID)
			if err != nil {
				return nil, fmt.Errorf("open blackboard: %w", err)
			}
			defer board.Close()

			if input.Certainty == "" {
				input.Certainty = blackboard.CertaintyConjecture
			}

			entry := &db.BlackboardEntry{
				ID:              uuid.New().String()[:8],
				ProjectID:       projectID,
				Category:        input.Category,
				Content:         input.Content,
				Certainty:       input.Certainty,
				Author:          agentID,
				ConfidenceScore: input.ConfidenceScore,
				Reasoning:       input.Reasoning,
			}

			if err := board.WriteEntrySync(entry); err != nil {
				return nil, fmt.Errorf("write entry: %w", err)
			}

			return &BlackboardWriteOutput{
				EntryID:  entry.ID,
				Category: input.Category,
				Message:  "Entry written to blackboard",
			}, nil
		},
	)
}

// --- Term Tool (Command Execution with Safety Review) ---

// TermInput is the input for the term tool
type TermInput struct {
	Command string `json:"command" jsonschema:"description=Shell command to execute,required"`
}

// TermOutput is the output for the term tool
type TermOutput struct {
	Output   string `json:"output"`
	IsSafe   bool   `json:"is_safe"`
	Rejected string `json:"rejected,omitempty"`
}

// NewTermTool creates an Eino tool for command execution with LLM safety review
func NewTermTool(agentID string, safetyLLM func(ctx context.Context, prompt string) (string, error)) (tool.InvokableTool, error) {
	return utils.InferTool(
		"term",
		"Execute a shell command. Each command is reviewed by an LLM safety checker before execution. Dangerous commands (rm -rf /, dd, format, etc.) will be rejected.",
		func(ctx context.Context, input TermInput) (*TermOutput, error) {
			// Safety review using LLM
			safetyPrompt := fmt.Sprintf(`你是一个命令行安全审查员。请评估以下命令是否安全执行。

危险命令示例（必须拒绝）：
- rm -rf /          → 删除整个文件系统
- rm -rf /*         → 同上
- dd if=/dev/zero   → 覆盖磁盘数据
- :(){ :|:& };:     → fork 炸弹
- chmod -R 777 /    → 破坏整个系统权限
- > /etc/passwd     → 清空系统用户文件

需要评估的命令: %s

请回答 SAFE 或 DANGEROUS，如果危险请说明原因和建议的安全替代方案。`, input.Command)

			review, err := safetyLLM(ctx, safetyPrompt)
			if err != nil {
				return nil, fmt.Errorf("safety review failed: %w", err)
			}

			// Check if the command is deemed dangerous
			if isDangerous(review) {
				return &TermOutput{
					IsSafe:   false,
					Rejected: review,
				}, nil
			}

			// Execute the command
			// Note: In production, use proper sandboxed execution
			return &TermOutput{
				Output: fmt.Sprintf("Command approved but execution not yet implemented in Phase 1. Command: %s", input.Command),
				IsSafe: true,
			}, nil
		},
	)
}

// isDangerous checks if the LLM safety review indicates a dangerous command
func isDangerous(review string) bool {
	// Simple heuristic: check if "DANGEROUS" appears in the response
	return len(review) >= 9 && (review[:9] == "DANGEROUS" || containsIgnoreCase(review, "DANGEROUS"))
}

func containsIgnoreCase(s, substr string) bool {
	slen := len(s)
	sublen := len(substr)
	if sublen > slen {
		return false
	}
	for i := 0; i <= slen-sublen; i++ {
		match := true
		for j := 0; j < sublen; j++ {
			sc := s[i+j]
			bc := substr[j]
			if sc >= 'A' && sc <= 'Z' {
				sc += 32
			}
			if bc >= 'A' && bc <= 'Z' {
				bc += 32
			}
			if sc != bc {
				match = false
				break
			}
		}
		if match {
			return true
		}
	}
	return false
}

// --- Memory Tool ---

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

// NewMemoryReadTool creates an Eino tool for reading personal memory
func NewMemoryReadTool(agentDataDir, agentID string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"memory_read",
		"Read your personal memory (memory.md). Contains declarative facts separated by § symbols.",
		func(ctx context.Context, input MemoryReadInput) (*MemoryReadOutput, error) {
			// TODO: Read from data/agents/{agentID}/memory.md
			return &MemoryReadOutput{
				Content: "", // Empty for Phase 1
			}, nil
		},
	)
}

// NewMemoryWriteTool creates an Eino tool for writing to personal memory
func NewMemoryWriteTool(agentDataDir, agentID string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"memory_write",
		"Append a declarative fact to your personal memory (memory.md). Facts are §-separated. Write only facts, not instructions.",
		func(ctx context.Context, input MemoryWriteInput) (*MemoryWriteOutput, error) {
			// TODO: Append to data/agents/{agentID}/memory.md
			return &MemoryWriteOutput{
				Message: "Fact saved to personal memory",
			}, nil
		},
	)
}

// --- Meeting Tool ---

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

// NewMeetingTool creates an Eino tool for meeting interactions
func NewMeetingTool(agentID, projectID string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"meeting",
		"Participate in meetings: speak (add your input), resolve (propose a resolution), or close (end the meeting).",
		func(ctx context.Context, input MeetingInput) (*MeetingOutput, error) {
			// TODO: Implement meeting system in Phase 3
			return &MeetingOutput{
				Message: fmt.Sprintf("Meeting action '%s' received. Meeting system will be implemented in Phase 3.", input.Action),
			}, nil
		},
	)
}
