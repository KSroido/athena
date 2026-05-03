package core

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"

	"github.com/ksroido/athena/internal/config"
)

// providerState tracks the runtime state of a single LLM provider
type providerState struct {
	config     config.LLMProviderConfig
	chatModel  *openai.ChatModel
	cooldownUntil time.Time  // 429 冷却截止时间
	failCount  int           // 连续失败次数
}

// LLMClient wraps multiple Eino ChatModels with fallback and 429 handling
type LLMClient struct {
	mu         sync.RWMutex
	providers  []*providerState
	maxRetries int
	cooldown   int // seconds
	modelName  string // primary model name (for display)
}

// NewLLMClient creates a new LLM client with provider fallback chain
func NewLLMClient(ctx context.Context, cfg *config.LLMConfig) (*LLMClient, error) {
	if cfg == nil {
		return nil, fmt.Errorf("llm config is required")
	}

	providers := cfg.GetProviders()
	if len(providers) == 0 {
		return nil, fmt.Errorf("no llm providers configured")
	}

	maxRetries := cfg.MaxRetries
	if maxRetries <= 0 {
		maxRetries = 2
	}
	cooldown := cfg.RetryCooldown
	if cooldown <= 0 {
		cooldown = 30
	}

	states := make([]*providerState, 0, len(providers))
	for _, p := range providers {
		cm, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
			BaseURL: p.BaseURL,
			APIKey:  p.APIKey,
			Model:   p.Model,
		})
		if err != nil {
			log.Printf("[llm] WARNING: failed to init provider %s/%s: %v", p.BaseURL, p.Model, err)
			continue
		}
		states = append(states, &providerState{
			config:    p,
			chatModel: cm,
		})
		log.Printf("[llm] initialized provider: %s/%s (weight=%d)", p.BaseURL, p.Model, p.Weight)
	}

	if len(states) == 0 {
		return nil, fmt.Errorf("no llm providers could be initialized")
	}

	return &LLMClient{
		providers:  states,
		maxRetries: maxRetries,
		cooldown:   cooldown,
		modelName:  states[0].config.Model,
	}, nil
}

// NewLLMClientSingle creates a single-provider client (legacy convenience)
func NewLLMClientSingle(ctx context.Context, baseURL, apiKey, model string) (*LLMClient, error) {
	cfg := &config.LLMConfig{
		Providers: []config.LLMProviderConfig{
			{BaseURL: baseURL, APIKey: apiKey, Model: model, Weight: 100},
		},
	}
	return NewLLMClient(ctx, cfg)
}

// Chat sends a chat completion request with automatic fallback on errors/429
func (c *LLMClient) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	var lastErr error

	for attempt := 0; attempt < c.maxRetries; attempt++ {
		provider := c.pickProvider()
		if provider == nil {
			// All providers in cooldown, wait for shortest cooldown
			waitDur := c.shortestCooldown()
			if waitDur > 0 {
				log.Printf("[llm] all providers in cooldown, waiting %v", waitDur)
				select {
				case <-time.After(waitDur):
					continue
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled while waiting for provider cooldown: %w", ctx.Err())
				}
			}
			return nil, fmt.Errorf("no available llm providers")
		}

		resp, err := provider.chatModel.Generate(ctx, messages)
		if err != nil {
			lastErr = err
			c.handleProviderError(provider, err)

			// If 429, immediately try next provider (don't waste retry count)
			if isRateLimitError(err) {
				log.Printf("[llm] 429 from %s/%s, cooling for %ds, trying next provider",
					provider.config.BaseURL, provider.config.Model, c.cooldown)
				continue
			}

			// For other errors, try next provider
			log.Printf("[llm] error from %s/%s: %v, trying next provider",
				provider.config.BaseURL, provider.config.Model, err)
			continue
		}

		// Success — reset fail count
		c.mu.Lock()
		provider.failCount = 0
		c.mu.Unlock()

		return resp, nil
	}

	return nil, fmt.Errorf("all %d attempts failed, last error: %w", c.maxRetries, lastErr)
}

// ChatWithSystem sends a chat with a system prompt
func (c *LLMClient) ChatWithSystem(ctx context.Context, systemPrompt string, userMessage string) (*schema.Message, error) {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	}
	return c.Chat(ctx, messages)
}

// ModelName returns the primary model name
func (c *LLMClient) ModelName() string {
	return c.modelName
}

// PrimaryChatModel returns the primary (highest-priority) Eino ChatModel.
// Used by agent_loop_v2.go for Eino's tool-binding ChatModel interface.
// For general chat requests, use Chat() or ChatWithSystem() which support fallback.
func (c *LLMClient) PrimaryChatModel() *openai.ChatModel {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if len(c.providers) > 0 {
		return c.providers[0].chatModel
	}
	return nil
}

// pickProvider selects the highest-priority available (not in cooldown) provider
func (c *LLMClient) pickProvider() *providerState {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	for _, p := range c.providers {
		if now.After(p.cooldownUntil) {
			return p
		}
	}
	return nil
}

// shortestCooldown returns the duration until the earliest provider comes back
func (c *LLMClient) shortestCooldown() time.Duration {
	c.mu.RLock()
	defer c.mu.RUnlock()

	now := time.Now()
	shortest := time.Duration(0)
	for _, p := range c.providers {
		remaining := p.cooldownUntil.Sub(now)
		if remaining > 0 {
			if shortest == 0 || remaining < shortest {
				shortest = remaining
			}
		}
	}
	return shortest
}

// handleProviderError processes errors and updates provider state
func (c *LLMClient) handleProviderError(p *providerState, err error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	p.failCount++

	if isRateLimitError(err) {
		// Parse Retry-After header if available, otherwise use configured cooldown
		cooldownDur := time.Duration(c.cooldown) * time.Second
		if retryAfter := parseRetryAfter(err); retryAfter > 0 {
			cooldownDur = retryAfter
		}
		p.cooldownUntil = time.Now().Add(cooldownDur)
		log.Printf("[llm] provider %s/%s rate-limited, cooldown until %s",
			p.config.BaseURL, p.config.Model, p.cooldownUntil.Format("15:04:05"))
	} else if p.failCount >= 3 {
		// After 3 consecutive non-429 failures, give a short cooldown
		p.cooldownUntil = time.Now().Add(10 * time.Second)
		log.Printf("[llm] provider %s/%s failed %d times, short cooldown until %s",
			p.config.BaseURL, p.config.Model, p.failCount, p.cooldownUntil.Format("15:04:05"))
	}
}

// isRateLimitError checks if the error is a 429 Too Many Requests
func isRateLimitError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "429") ||
		strings.Contains(errStr, "rate_limit") ||
		strings.Contains(errStr, "rate limit") ||
		strings.Contains(errStr, "too many requests") ||
		strings.Contains(errStr, "quota exceeded")
}

// parseRetryAfter attempts to extract Retry-After duration from error
func parseRetryAfter(err error) time.Duration {
	if err == nil {
		return 0
	}
	errStr := err.Error()

	// Try to find "retry after Xs" pattern
	for _, prefix := range []string{"retry after ", "Retry-After: ", "retry-after: "} {
		if idx := strings.Index(strings.ToLower(errStr), strings.ToLower(prefix)); idx >= 0 {
			rest := errStr[idx+len(prefix):]
			// Extract number
			numEnd := 0
			for numEnd < len(rest) && (rest[numEnd] >= '0' && rest[numEnd] <= '9') {
				numEnd++
			}
			if numEnd > 0 {
				seconds, err := strconv.Atoi(rest[:numEnd])
				if err == nil && seconds > 0 {
					return time.Duration(seconds) * time.Second
				}
			}
		}
	}

	// Check for HTTP status code 429 with Retry-After header
	if strings.Contains(errStr, "429") {
		// Try to find header value
		for _, hdr := range []string{"Retry-After:", "retry-after:"} {
			if idx := strings.Index(errStr, hdr); idx >= 0 {
				rest := errStr[idx+len(hdr):]
				rest = strings.TrimSpace(rest)
				numEnd := 0
				for numEnd < len(rest) && (rest[numEnd] >= '0' && rest[numEnd] <= '9') {
					numEnd++
				}
				if numEnd > 0 {
					seconds, err := strconv.Atoi(rest[:numEnd])
					if err == nil && seconds > 0 {
						return time.Duration(seconds) * time.Second
					}
				}
			}
		}
	}

	_ = http.StatusTooManyRequests // reference to suppress lint
	return 0
}
