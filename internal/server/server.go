package server

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"

	"github.com/ksroido/athena/internal/api"
	"github.com/ksroido/athena/internal/config"
	"github.com/ksroido/athena/internal/core"
	"github.com/ksroido/athena/internal/db"
)

// Server is the Athena HTTP server
type Server struct {
	cfg         *config.Config
	mainDB      *db.DB
	agentServer *core.AgentServer
	supervisor  *core.Supervisor
	engine      *gin.Engine
}

// New creates a new Athena server
func New(cfg *config.Config) (*Server, error) {
	// Initialize main database
	dataDir := cfg.Agents.DataDir
	if dataDir == "" {
		dataDir = "./data"
	}
	mainDB, err := db.New(dataDir)
	if err != nil {
		return nil, fmt.Errorf("init database: %w", err)
	}

	// Create LLM client
	llm, err := core.NewLLMClient(nil, cfg.LLM.BaseURL, cfg.LLM.APIKey, cfg.LLM.Model)
	if err != nil {
		return nil, fmt.Errorf("init LLM client: %w", err)
	}

	// Create supervisor
	supervisor := core.NewSupervisor(
		"athena-agent",
		core.LLMConfig{
			BaseURL: cfg.LLM.BaseURL,
			APIKey:  cfg.LLM.APIKey,
			Model:   cfg.LLM.Model,
		},
		mainDB,
	)

	// Create AgentServer
	agentServer := core.NewAgentServer(llm, mainDB, supervisor)

	// Setup Gin
	if os.Getenv("GIN_MODE") == "" {
		gin.SetMode(gin.DebugMode)
	}
	engine := gin.Default()

	s := &Server{
		cfg:         cfg,
		mainDB:      mainDB,
		agentServer: agentServer,
		supervisor:  supervisor,
		engine:      engine,
	}

	// Register routes
	s.registerRoutes()

	return s, nil
}

// registerRoutes sets up all API routes
func (s *Server) registerRoutes() {
	// Health check
	s.engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Serve frontend static files
	s.engine.Static("/assets", "./frontend/dist/assets")
	s.engine.StaticFile("/favicon.ico", "./frontend/dist/favicon.ico")

	// API v1 group
	v1 := s.engine.Group("/api")
	{
		// CEO chat endpoint (AgentServer)
		v1.POST("/chat", api.HandleChat(s.agentServer))

		// Projects
		v1.GET("/projects", api.HandleListProjects(s.mainDB))
		v1.POST("/projects", api.HandleCreateProject(s.mainDB))
		v1.GET("/projects/:id", api.HandleGetProject(s.mainDB))

		// Blackboard
		v1.GET("/projects/:id/blackboard", api.HandleGetBlackboard(s.mainDB))
		v1.POST("/projects/:id/blackboard", api.HandleWriteBlackboard(s.mainDB))

		// Agents
		v1.GET("/agents", api.HandleListAgents(s.supervisor))
		v1.GET("/projects/:id/agents", api.HandleListProjectAgents(s.mainDB))

		// Meetings
		v1.GET("/projects/:id/meetings", api.HandleListMeetings(s.mainDB))
	}

	// SPA fallback: serve index.html for all non-API, non-static routes
	s.engine.NoRoute(func(c *gin.Context) {
		// Don't serve index.html for API routes
		if len(c.Request.URL.Path) >= 4 && c.Request.URL.Path[:4] == "/api" {
			c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
			return
		}
		c.File("./frontend/dist/index.html")
	})
}

// Run starts the Athena server
func (s *Server) Run() error {
	addr := fmt.Sprintf("%s:%d", s.cfg.Server.Host, s.cfg.Server.Port)

	log.Printf("Athena server starting on %s", addr)
	log.Printf("Frontend: http://localhost:%d", s.cfg.Server.Port)

	// Graceful shutdown
	go func() {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
		<-quit
		log.Println("Shutting down Athena server...")
		s.mainDB.Close()
		os.Exit(0)
	}()

	return s.engine.Run(addr)
}

// Close cleans up resources
func (s *Server) Close() {
	if s.mainDB != nil {
		s.mainDB.Close()
	}
}
