package llm

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client       *openai.Client
	model        string
	systemPrompt string
}

type ClientConfig struct {
	APIKey       string
	BaseURL      string // optional: defaults to OpenAI; set for OpenRouter, Ollama, LiteLLM, etc.
	Model        string // optional: defaults to gpt-4o-mini
	SystemPrompt string
}

type headerTransport struct {
	base    http.RoundTripper
	headers map[string]string
}

func (t *headerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for k, v := range t.headers {
		req.Header.Set(k, v)
	}
	return t.base.RoundTrip(req)
}

func NewOpenAIClient(cfg ClientConfig) *OpenAIClient {
	model := cfg.Model
	if model == "" {
		model = openai.GPT4oMini
	}

	var client *openai.Client
	if cfg.BaseURL != "" {
		config := openai.DefaultConfig(cfg.APIKey)
		config.BaseURL = cfg.BaseURL
		config.HTTPClient = &http.Client{
			Transport: &headerTransport{
				base: http.DefaultTransport,
				headers: map[string]string{
					"HTTP-Referer": "https://github.com/contract-scanner-ia",
					"X-Title":     "Contract Scanner",
				},
			},
		}
		client = openai.NewClientWithConfig(config)
	} else {
		client = openai.NewClient(cfg.APIKey)
	}

	return &OpenAIClient{
		client:       client,
		model:        model,
		systemPrompt: cfg.SystemPrompt,
	}
}

var mdJSONFence = regexp.MustCompile("(?s)^```(?:json)?\\s*(\\{.*\\})\\s*```$")

func stripMarkdownJSON(s string) string {
	s = strings.TrimSpace(s)
	if m := mdJSONFence.FindStringSubmatch(s); len(m) == 2 {
		return strings.TrimSpace(m[1])
	}
	return s
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

	content := stripMarkdownJSON(resp.Choices[0].Message.Content)

	if !json.Valid([]byte(content)) {
		return nil, fmt.Errorf("openai returned invalid JSON: %s", content)
	}

	return json.RawMessage(content), nil
}
