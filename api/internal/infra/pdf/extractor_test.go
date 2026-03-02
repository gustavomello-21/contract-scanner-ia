package pdf

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestCleanAndNormalize_RemovesD4SignNoise(t *testing.T) {
	raw := "D4Sign\nDocumento assinado eletronicamente, conforme MP 2.200-2/01, Art. 10º, §2.\n" +
		"Cláusula 1 - Objeto do contrato\x00\n" +
		"Cláusula 1 - Objeto do contrato\n"

	cleaned, stats := cleanAndNormalize(raw)

	if strings.Contains(strings.ToLower(cleaned), "d4sign") {
		t.Fatalf("expected watermark to be removed, got: %q", cleaned)
	}
	if strings.Contains(cleaned, "\x00") {
		t.Fatalf("expected control chars removed, got: %q", cleaned)
	}
	if !strings.Contains(cleaned, "Cláusula 1 - Objeto do contrato") {
		t.Fatalf("expected contractual content preserved, got: %q", cleaned)
	}
	if stats.removedWatermark == 0 {
		t.Fatal("expected watermark removals to be tracked")
	}
}

func TestScoreQuality_RepetitiveVsContractualText(t *testing.T) {
	repetitive := strings.Repeat("D4Sign\n", 200)
	good := strings.Repeat("CONTRATO DE PRESTACAO DE SERVICOS\nCLAUSULA DE OBJETO E RESPONSABILIDADE\nPRAZO, PAGAMENTO, MULTA E RESCISAO\nPARTES ACORDAM A VIGENCIA DE 12 MESES\n", 20)

	low := scoreQuality(repetitive)
	high := scoreQuality(good)

	if !(low < high) {
		t.Fatalf("expected repetitive text score < good text score, low=%.2f high=%.2f", low, high)
	}
	if high < 0.35 {
		t.Fatalf("expected good text to have acceptable quality score, got %.2f", high)
	}
}

func TestExtract_UsesOCRWhenNativeIsPoor(t *testing.T) {
	extractor := NewHybridExtractor(Config{
		OCREnabled:      true,
		OCRLang:         "por+eng",
		OCRDPI:          300,
		MinQualityScore: 0.35,
		OCRMaxPages:     10,
	})

	extractor.nativeExtract = func(string) (string, error) {
		return strings.Repeat("D4Sign\nDocumento assinado eletronicamente\n", 50), nil
	}
	extractor.ocrExtractWith = func(context.Context, string) (string, error) {
		return strings.Repeat("CONTRATO\nCLAUSULA DE OBJETO E PAGAMENTO\nPARTES E PRAZO DE VIGENCIA\n", 30), nil
	}

	result, err := extractor.Extract(context.Background(), "/tmp/fake.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != sourceOCR {
		t.Fatalf("expected source %q, got %q", sourceOCR, result.Source)
	}
	if result.QualityScore < 0.35 {
		t.Fatalf("expected acceptable quality score, got %.2f", result.QualityScore)
	}
}

func TestExtract_FallsBackToNativeWhenOCRFails(t *testing.T) {
	extractor := NewHybridExtractor(Config{
		OCREnabled:      true,
		OCRLang:         "por+eng",
		OCRDPI:          300,
		MinQualityScore: 0.95,
		OCRMaxPages:     10,
	})

	extractor.nativeExtract = func(string) (string, error) {
		return strings.Repeat("CONTRATO DE SERVICOS\nCLAUSULA DE OBJETO\nPRAZO E RESPONSABILIDADE\n", 20), nil
	}
	extractor.ocrExtractWith = func(context.Context, string) (string, error) {
		return "", errors.New("tesseract not available")
	}

	result, err := extractor.Extract(context.Background(), "/tmp/fake.pdf")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Source != sourceNative {
		t.Fatalf("expected source %q, got %q", sourceNative, result.Source)
	}
	if len(result.Warnings) == 0 {
		t.Fatal("expected warnings when OCR fails")
	}
}
