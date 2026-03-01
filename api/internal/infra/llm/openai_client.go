package llm

import (
	"context"
	"encoding/json"
	"fmt"

	openai "github.com/sashabaranov/go-openai"
)

type OpenAIClient struct {
	client *openai.Client
	model  string
}

func NewOpenAIClient(apiKey string) *OpenAIClient {
	return &OpenAIClient{
		client: openai.NewClient(apiKey),
		model:  openai.GPT4oMini,
	}
}

const systemPrompt = `Você é um assistente jurídico especializado em análise de contratos.
Analise o contrato fornecido e retorne um JSON com a seguinte estrutura:
{
  "summary": "resumo do contrato em 2-3 frases",
  "parties": ["parte 1", "parte 2"],
  "contract_type": "tipo do contrato (ex: prestação de serviços, locação, compra e venda)",
  "key_clauses": [
    {"clause": "nome da cláusula", "description": "descrição breve", "risk_level": "low|medium|high"}
  ],
  "risks": ["risco 1", "risco 2"],
  "recommendations": ["recomendação 1", "recomendação 2"],
  "overall_risk": "low|medium|high"
}
Retorne APENAS o JSON, sem markdown ou texto adicional.`

func (o *OpenAIClient) AnalyzeContract(ctx context.Context, text string) (json.RawMessage, error) {
	resp, err := o.client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
		Model: o.model,
		Messages: []openai.ChatCompletionMessage{
			{Role: openai.ChatMessageRoleSystem, Content: systemPrompt},
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
