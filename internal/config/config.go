package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config holds the Athena server configuration
type Config struct {
	Server    ServerConfig     `yaml:"server"`
	LLM       LLMConfig        `yaml:"llm"`
	Company   CompanyConfig    `yaml:"company"`
	Blackboard BlackboardConfig `yaml:"blackboard"`
	Meeting   MeetingConfig    `yaml:"meeting"`
	Agents    AgentsConfig     `yaml:"agents"`
	Logging   LoggingConfig    `yaml:"logging"`
}

type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

type LLMProviderConfig struct {
	BaseURL string `yaml:"base_url"`
	APIKey  string `yaml:"api_key"`
	Model   string `yaml:"model"`
	Weight  int    `yaml:"weight"`  // 优先级权重，越大越优先
}

type LLMConfig struct {
	BaseURL   string             `yaml:"base_url"`
	APIKey    string             `yaml:"api_key"`
	Model     string             `yaml:"model"`
	Providers []LLMProviderConfig `yaml:"providers"`
	// Fallback settings
	MaxRetries    int `yaml:"max_retries"`     // 单 provider 最大重试次数，默认 2
	RetryCooldown int `yaml:"retry_cooldown"`  // 429 后冷却秒数，默认 30
}

// GetProviders returns the provider chain in priority order (highest weight first)
func (l *LLMConfig) GetProviders() []LLMProviderConfig {
	if len(l.Providers) == 0 {
		// Legacy mode: single provider from top-level fields
		return []LLMProviderConfig{{
			BaseURL: l.BaseURL,
			APIKey:  l.APIKey,
			Model:   l.Model,
			Weight:  100,
		}}
	}
	// Sort by weight descending
	sorted := make([]LLMProviderConfig, len(l.Providers))
	copy(sorted, l.Providers)
	for i := 0; i < len(sorted)-1; i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[j].Weight > sorted[i].Weight {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}
	return sorted
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
			MaxRetries:    2,
			RetryCooldown: 30,
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

	// Expand environment variables in API keys
	cfg.expandEnv()

	return cfg, nil
}

// expandEnv replaces $ENV_VAR or ${ENV_VAR} in API keys
func (c *Config) expandEnv() {
	c.LLM.APIKey = expandEnvStr(c.LLM.APIKey)
	for i := range c.LLM.Providers {
		c.LLM.Providers[i].APIKey = expandEnvStr(c.LLM.Providers[i].APIKey)
	}
}

func expandEnvStr(s string) string {
	if len(s) > 0 && (s[0] == '$' || (len(s) > 1 && s[:2] == "${")) {
		return os.ExpandEnv(s)
	}
	return s
}

// Validate checks the configuration for required fields
func (c *Config) Validate() error {
	providers := c.LLM.GetProviders()
	if len(providers) == 0 {
		return fmt.Errorf("at least one llm provider is required")
	}
	for i, p := range providers {
		if p.BaseURL == "" {
			return fmt.Errorf("llm.providers[%d].base_url is required", i)
		}
		if p.APIKey == "" {
			return fmt.Errorf("llm.providers[%d].api_key is required", i)
		}
		if p.Model == "" {
			return fmt.Errorf("llm.providers[%d].model is required", i)
		}
	}
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("server.port must be between 1 and 65535")
	}
	return nil
}
