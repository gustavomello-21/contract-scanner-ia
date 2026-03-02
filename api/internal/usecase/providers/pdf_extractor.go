package providers

import "context"

type ExtractResult struct {
	Text         string
	Source       string
	QualityScore float64
	Warnings     []string
}

type PDFExtractor interface {
	Extract(ctx context.Context, filePath string) (*ExtractResult, error)
}
