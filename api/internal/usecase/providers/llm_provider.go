package providers

import (
	"context"
	"encoding/json"
)

type LLMProvider interface {
	AnalyzeContract(ctx context.Context, text string) (json.RawMessage, error)
}
