package db

import (
	"database/sql"
	"time"
)

// Project represents a project in the system
type Project struct {
	ID                  string         `json:"id"`
	Name                string         `json:"name"`
	Description         string         `json:"description,omitempty"`
	OriginalRequirement string         `json:"original_requirement,omitempty"`
	RequirementSummary  string         `json:"requirement_summary,omitempty"`
	Status              string         `json:"status"`
	Priority            int            `json:"priority"`
	CreatedAt           time.Time      `json:"created_at"`
	UpdatedAt           time.Time      `json:"updated_at"`
}

// Agent represents an agent in the system
type Agent struct {
	ID         string         `json:"id"`
	Name       string         `json:"name"`
	Role       string         `json:"role"`
	Status     string         `json:"status"`
	Tools      string         `json:"tools,omitempty"`       // JSON
	MCPServers string         `json:"mcp_servers,omitempty"` // JSON
	Model      string         `json:"model"`
	State      string         `json:"state,omitempty"`       // JSON
	CreatedBy  string         `json:"created_by,omitempty"`
	CreatedAt  time.Time      `json:"created_at"`
}

// AgentTask represents a task assigned to an agent
type AgentTask struct {
	ID             string         `json:"id"`
	ProjectID      string         `json:"project_id"`
	AgentID        string         `json:"agent_id"`
	Title          string         `json:"title"`
	Description    string         `json:"description,omitempty"`
	Status         string         `json:"status"`
	Priority       int            `json:"priority"`
	Result         string         `json:"result,omitempty"`
	ReviewStatus   sql.NullString `json:"review_status"`
	ReviewedBy     sql.NullString `json:"reviewed_by"`
	DependsOn      sql.NullString `json:"depends_on"`
	IdempotencyKey sql.NullString `json:"idempotency_key"`
	CreatedAt      time.Time      `json:"created_at"`
	CompletedAt    sql.NullTime   `json:"completed_at"`
}

// ProjectMember represents an agent's membership in a project
type ProjectMember struct {
	ID        string    `json:"id"`
	ProjectID string    `json:"project_id"`
	AgentID   string    `json:"agent_id"`
	Role      string    `json:"role"`
	JoinedAt  time.Time `json:"joined_at"`
}

// BlackboardEntry represents a blackboard entry (used in project-specific DBs)
type BlackboardEntry struct {
	ID              string    `json:"id"`
	ProjectID       string    `json:"project_id"`
	Category        string    `json:"category"`
	Content         string    `json:"content"`
	Certainty       string    `json:"certainty"`
	Author          string    `json:"author,omitempty"`
	ConfidenceScore *int      `json:"confidence_score,omitempty"`
	Reasoning       string    `json:"reasoning,omitempty"`
	LastVerifiedAt  *time.Time `json:"last_verified_at,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
}

// Meeting represents a meeting
type Meeting struct {
	ID                 string     `json:"id"`
	ProjectID          string     `json:"project_id"`
	ConvenerID         string     `json:"convener_id"`
	Status             string     `json:"status"`
	Resolution         string     `json:"resolution,omitempty"`
	ArchivedDiscussion string     `json:"archived_discussion,omitempty"`
	CreatedAt          time.Time  `json:"created_at"`
	ClosedAt           *time.Time `json:"closed_at,omitempty"`
}
