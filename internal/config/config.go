package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the Athena server configuration
type Config struct {
	Server    ServerConfig    `yaml:"server"`
	LLM       LLMConfig       `yaml:"llm"`
	Company   CompanyConfig   `yaml:"company"`
	Blackboard BlackboardConfig `yaml:"blackboard"`
	Meeting   MeetingConfig   `yaml:"meeting"`
	Agents    AgentsConfig    `yaml:"agents"`
	Logging   LoggingConfig   `yaml:"logging"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LLMConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
}

type CompanyConfig struct {
	MaxAgents   int `yaml:"max_agents"`
	MaxMemoryMB int `yaml:"max_memory_mb"`
}

type BlackboardConfig struct {
	DataDir string `yaml:"data_dir"`
}

type MeetingConfig struct {
	DataDir string `yaml:"data_dir"`
}

type AgentsConfig struct {
	DataDir string `yaml:"data_dir"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns a configuration with default values
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			Host: "0.0.0.0",
			Port: 8080,
		},
		LLM: LLMConfig{
			BaseURL: "https://api.openai.com/v1",
			Model:   "gpt-4o",
		},
		Company: CompanyConfig{
			MaxAgents:   100,
			MaxMemoryMB: 16384,
		},
		Blackboard: BlackboardConfig{
			DataDir: "./data/board",
		},
		Meeting: MeetingConfig{
			DataDir: "./data/meetings",
		},
		Agents: AgentsConfig{
			DataDir: "./data/agents",
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "./data/logs/athena.log",
		},
	}
}

// Load reads configuration from a YAML file
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config file: %w", err)
	}

	cfg := DefaultConfig()
	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parse config: %w", err)
	}

	return cfg, nil
}

// Validate checks the configuration for required fields
func (c *Config) Validate() error {
	if c.LLM.BaseURL == "" {
		return fmt.Errorf("llm.base_url is required")
	}
	if c.LLM.APIKey == "" {
		return fmt.Errorf("llm.api_key is required")
	}
	if c.LLM.Model == "" {
		return fmt.Errorf("llm.model is required")
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	return nil
}
