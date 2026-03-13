package llm

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client       *openai.Client
	model        string
	systemPrompt string
}

func NewOpenAIClient(apiKey string, systemPrompt string) *OpenAIClient {
	return &OpenAIClient{
		client:       openai.NewClient(apiKey),
		model:        openai.GPT4oMini,
		systemPrompt: systemPrompt,
	}
}

func (o *OpenAIClient) AnalyzeContract(ctx context.Context, text string) (json.RawMessage, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: o.systemPrompt},
			{Role: openai.ChatMessageRoleUser, Content: text},
		},
		Temperature: 0.2,
	})
	if err != nil {
		return nil, fmt.Errorf("openai request failed: %w", err)
	}

	content := resp.Choices[0].Message.Content

	if !json.Valid([]byte(content)) {
		return nil, fmt.Errorf("openai returned invalid JSON: %s", content)
	}

	return json.RawMessage(content), nil
}
