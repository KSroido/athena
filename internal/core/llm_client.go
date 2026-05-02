package core

import (
	"context"
	"fmt"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/schema"
)

// LLMClient wraps Eino's ChatModel for LLM calls
type LLMClient struct {
	chatModel *openai.ChatModel
	modelName string
}

// NewLLMClient creates a new LLM client using Eino's OpenAI-compatible ChatModel
func NewLLMClient(ctx context.Context, baseURL, apiKey, model string) (*LLMClient, error) {
	chatModel, err := openai.NewChatModel(ctx, &openai.ChatModelConfig{
		BaseURL: baseURL,
		APIKey:  apiKey,
		Model:   model,
	})
	if err != nil {
		return nil, fmt.Errorf("create chat model: %w", err)
	}

	return &LLMClient{
		chatModel: chatModel,
		modelName: model,
	}, nil
}

// Chat sends a chat completion request and returns the response
func (c *LLMClient) Chat(ctx context.Context, messages []*schema.Message) (*schema.Message, error) {
	resp, err := c.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, fmt.Errorf("chat generate: %w", err)
	}
	return resp, nil
}

// ChatWithSystem sends a chat with a system prompt
func (c *LLMClient) ChatWithSystem(ctx context.Context, systemPrompt string, userMessage string) (*schema.Message, error) {
	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userMessage),
	}
	return c.Chat(ctx, messages)
}

// ModelName returns the configured model name
func (c *LLMClient) ModelName() string {
	return c.modelName
}
