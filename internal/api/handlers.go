package api

import (
	"database/sql"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/ksroido/athena/internal/blackboard"
	"github.com/ksroido/athena/internal/core"
	"github.com/ksroido/athena/internal/db"
)

// --- CEO Chat ---

// ChatRequest is the request body for the CEO chat endpoint
type ChatRequest struct {
	Message string `json:"message" binding:"required"`
}

// HandleChat processes CEO messages through AgentServer
func HandleChat(agentServer *core.AgentServer) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req ChatRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "message is required"})
			return
		}

		resp, err := agentServer.ProcessCEOMessage(c.Request.Context(), req.Message)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"response": resp})
	}
}

// --- Projects ---

// HandleListProjects returns all projects
func HandleListProjects(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		rows, err := mainDB.DB().Query(
			"SELECT id, name, status, priority, original_requirement, created_at, updated_at FROM projects ORDER BY created_at DESC",
		)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var projects []map[string]interface{}
		for rows.Next() {
			var id, name, status, originalReq string
			var priority int
			var createdAt, updatedAt time.Time
			if err := rows.Scan(&id, &name, &status, &priority, &originalReq, &createdAt, &updatedAt); err != nil {
				continue
			}
			projects = append(projects, map[string]interface{}{
				"id":                   id,
				"name":                 name,
				"status":               status,
				"priority":             priority,
				"original_requirement": originalReq,
				"created_at":           createdAt,
				"updated_at":           updatedAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"projects": projects})
	}
}

// CreateProjectRequest is the request body for creating a project
type CreateProjectRequest struct {
	Name                string `json:"name" binding:"required"`
	OriginalRequirement string `json:"original_requirement" binding:"required"`
	Description         string `json:"description"`
	Priority            int    `json:"priority"`
}

// HandleCreateProject creates a new project
func HandleCreateProject(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req CreateProjectRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "name and original_requirement are required"})
			return
		}

		projectID := uuid.New().String()[:8]
		if req.Priority == 0 {
			req.Priority = 5
		}

		_, err := mainDB.DB().Exec(`
			INSERT INTO projects (id, name, description, original_requirement, status, priority)
			VALUES (?, ?, ?, ?, 'active', ?)
		`, projectID, req.Name, req.Description, req.OriginalRequirement, req.Priority)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		// Create blackboard for this project
		board, err := blackboard.OpenBoard(mainDB.DataDir(), projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("create blackboard: %v", err)})
			return
		}

		// Write initial requirement as a goal
		board.WriteEntrySync(&db.BlackboardEntry{
			ID:        uuid.New().String()[:8],
			ProjectID: projectID,
			Category:  blackboard.CategoryGoal,
			Content:   req.OriginalRequirement,
			Certainty: blackboard.CertaintyCertain,
			Author:    "ceo",
		})
		board.Close()

		c.JSON(http.StatusCreated, gin.H{
			"id":      projectID,
			"name":    req.Name,
			"status":  "active",
			"message": "项目已创建，黑板已初始化",
		})
	}
}

// HandleGetProject returns a single project
func HandleGetProject(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		var p db.Project
		err := mainDB.DB().QueryRow(
			"SELECT id, name, description, status, priority, original_requirement, requirement_summary, created_at, updated_at FROM projects WHERE id = ?",
			projectID,
		).Scan(&p.ID, &p.Name, &p.Description, &p.Status, &p.Priority, &p.OriginalRequirement, &p.RequirementSummary, &p.CreatedAt, &p.UpdatedAt)
		if err == sql.ErrNoRows {
			c.JSON(http.StatusNotFound, gin.H{"error": "project not found"})
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, p)
	}
}

// --- Blackboard ---

// HandleGetBlackboard returns blackboard entries for a project
func HandleGetBlackboard(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		category := c.Query("category")

		board, err := blackboard.OpenBoard(mainDB.DataDir(), projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("open blackboard: %v", err)})
			return
		}
		defer board.Close()

		entries, err := board.ReadEntries(category, 100, 0)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusOK, gin.H{"entries": entries})
	}
}

// WriteBlackboardRequest is the request body for writing to the blackboard
type WriteBlackboardRequest struct {
	Category        string `json:"category" binding:"required"`
	Content         string `json:"content" binding:"required"`
	Certainty       string `json:"certainty"`
	Author          string `json:"author"`
	ConfidenceScore *int   `json:"confidence_score"`
	Reasoning       string `json:"reasoning"`
}

// HandleWriteBlackboard writes an entry to a project's blackboard
func HandleWriteBlackboard(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		var req WriteBlackboardRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "category and content are required"})
			return
		}

		if req.Certainty == "" {
			req.Certainty = blackboard.CertaintyConjecture
		}

		board, err := blackboard.OpenBoard(mainDB.DataDir(), projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("open blackboard: %v", err)})
			return
		}
		defer board.Close()

		entry := &db.BlackboardEntry{
			ID:              uuid.New().String()[:8],
			ProjectID:       projectID,
			Category:        req.Category,
			Content:         req.Content,
			Certainty:       req.Certainty,
			Author:          req.Author,
			ConfidenceScore: req.ConfidenceScore,
			Reasoning:       req.Reasoning,
		}

		if err := board.WriteEntrySync(entry); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		c.JSON(http.StatusCreated, gin.H{"entry": entry})
	}
}

// --- Agents ---

// HandleListAgents returns all running agents
func HandleListAgents(supervisor *core.Supervisor) gin.HandlerFunc {
	return func(c *gin.Context) {
		agents := supervisor.ListAgents()
		c.JSON(http.StatusOK, gin.H{"agents": agents})
	}
}

// HandleListProjectAgents returns agents for a specific project
func HandleListProjectAgents(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")

		rows, err := mainDB.DB().Query(`
			SELECT a.id, a.name, a.role, a.status, a.model, a.created_at
			FROM agents a
			JOIN project_members pm ON a.id = pm.agent_id
			WHERE pm.project_id = ?
			ORDER BY a.created_at
		`, projectID)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer rows.Close()

		var agents []map[string]interface{}
		for rows.Next() {
			var id, name, role, status, model string
			var createdAt time.Time
			if err := rows.Scan(&id, &name, &role, &status, &model, &createdAt); err != nil {
				continue
			}
			agents = append(agents, map[string]interface{}{
				"id":         id,
				"name":       name,
				"role":       role,
				"status":     status,
				"model":      model,
				"created_at": createdAt,
			})
		}

		c.JSON(http.StatusOK, gin.H{"agents": agents})
	}
}

// --- Meetings (placeholder) ---

// HandleListMeetings returns meetings for a project (placeholder)
func HandleListMeetings(mainDB *db.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		projectID := c.Param("id")
		c.JSON(http.StatusOK, gin.H{
			"project_id": projectID,
			"meetings":   []interface{}{},
			"message":    "会议系统待Phase 3实现",
		})
	}
}
