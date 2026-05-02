package main

import (
	"log"
	"os"

	"github.com/ksroido/athena/internal/config"
	"github.com/ksroido/athena/internal/server"
)

func main() {
	cfgPath := "config/athena.yaml"
	if len(os.Args) > 1 {
		cfgPath = os.Args[1]
	}

	cfg, err := loadConfig(cfgPath)
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	if err := cfg.Validate(); err != nil {
		log.Fatalf("Invalid configuration: %v", err)
	}

	srv, err := server.New(cfg)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}
	defer srv.Close()

	log.Println("Athena server starting...")
	if err := srv.Run(); err != nil {
		log.Fatalf("Server error: %v", err)
	}
}

func loadConfig(path string) (*config.Config, error) {
	if _, err := os.Stat(path); err == nil {
		return config.Load(path)
	}

	cfg := config.DefaultConfig()
	if v := os.Getenv("ATHENA_LLM_BASE_URL"); v != "" {
		cfg.LLM.BaseURL = v
	}
	if v := os.Getenv("ATHENA_LLM_API_KEY"); v != "" {
		cfg.LLM.APIKey = v
	}
	if v := os.Getenv("ATHENA_LLM_MODEL"); v != "" {
		cfg.LLM.Model = v
	}
	return cfg, nil
}
