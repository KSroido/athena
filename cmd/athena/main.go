package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/ksroido/athena/internal/config"
	"github.com/ksroido/athena/internal/server"
)

func main() {
	configPath := flag.String("config", "config/athena.yaml", "path to configuration file")
	flag.Parse()

	// Load configuration
	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	// Create and run server
	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	log.Println("=====================================")
	log.Println("  Athena — AI Agent 编排系统")
	log.Println("  像IT公司一样运作")
	log.Println("=====================================")

	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

// loadConfig loads configuration from file or environment
func loadConfig(path string) (*config.Config, error) {
	// Try to load from file
	if _, err := os.Stat(path); err == nil {
		cfg, err := config.Load(path)
		if err != nil {
			return nil, fmt.Errorf("load config from %s: %w", path, err)
		}
		return cfg, nil
	}

	// Fall back to environment variables (single provider mode)
	cfg := config.DefaultConfig()

	if baseURL := os.Getenv("ATHENA_LLM_BASE_URL"); baseURL != "" {
		cfg.LLM.BaseURL = baseURL
	}
	if apiKey := os.Getenv("ATHENA_LLM_API_KEY"); apiKey != "" {
		cfg.LLM.APIKey = apiKey
	}
	if model := os.Getenv("ATHENA_LLM_MODEL"); model != "" {
		cfg.LLM.Model = model
	}
	if port := os.Getenv("ATHENA_PORT"); port != "" {
		fmt.Sscanf(port, "%d", &cfg.Server.Port)
	}

	return cfg, nil
}
