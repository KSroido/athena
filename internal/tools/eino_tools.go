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

type BlackboardReadInput struct {
	Category string `json:"category" jsonschema:"description=Filter by category (goal/fact/discovery/decision/progress/resolution/auxiliary). Empty for all."`
	Limit    int    `json:"limit" jsonschema:"description=Maximum entries to return. Default 50."`
	Query    string `json:"query" jsonschema:"description=FTS5 full-text search query. If provided, category filter is ignored."`
}

type BlackboardReadOutput struct {
	Entries []*db.BlackboardEntry `json:"entries"`
	Count   int                   `json:"count"`
}

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
				entries, err = board.Search(input.Query, limit)
				if err != nil {
					return nil, fmt.Errorf("search: %w", err)
				}
			} else {
				entries, err = board.ReadEntries(input.Category, limit, 0)
				if err != nil {
					return nil, fmt.Errorf("read entries: %w", err)
				}
			}

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

type BlackboardWriteInput struct {
	Category        string `json:"category" jsonschema:"description=Entry category (fact/discovery/progress/decision/auxiliary),required"`
	Content         string `json:"content" jsonschema:"description=Entry content,required"`
	Certainty       string `json:"certainty" jsonschema:"description=Certainty level: certain/conjecture/pending_verification. Default: conjecture"`
	ConfidenceScore *int   `json:"confidence_score" jsonschema:"description=Self-assessment score 0-10. 10 requires reasoning."`
	Reasoning       string `json:"reasoning" jsonschema:"description=Reasoning chain for confidence_score=10 entries"`
}

type BlackboardWriteOutput struct {
	EntryID  string `json:"entry_id"`
	Category string `json:"category"`
	Message  string `json:"message"`
}

func NewBlackboardWriteTool(dataDir, projectID, agentID, role string) (tool.InvokableTool, error) {
	return utils.InferTool(
		"blackboard_write",
		"Write an entry to the project blackboard. Only write entries within your role's permissions. Mark certainty carefully: 'certain' only for 100% verified facts.",
		func(ctx context.Context, input BlackboardWriteInput) (*BlackboardWriteOutput, error) {
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
