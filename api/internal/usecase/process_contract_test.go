package usecase

import (
	"encoding/json"
	"testing"

	"contract-scanner/internal/usecase/providers"
)

func TestEnrichAnalysisResult_AddsWarningsConfidenceAndMissingInformation(t *testing.T) {
	raw := json.RawMessage(`{"summary":"ok","parties":[],"contract_type":"servicos","key_clauses":[],"risks":[],"recommendations":[],"overall_risk":"low"}`)
	extractResult := &providers.ExtractResult{
		Text:         "texto",
		Source:       "native",
		QualityScore: 0.2,
		Warnings:     []string{"native extraction failed"},
	}

	enriched, err := enrichAnalysisResult(raw, extractResult)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := map[string]any{}
	if err := json.Unmarshal(enriched, &payload); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	warnings, ok := payload["analysis_warnings"].([]any)
	if !ok || len(warnings) == 0 {
		t.Fatalf("expected analysis_warnings to be populated, got: %#v", payload["analysis_warnings"])
	}
	if _, ok := payload["confidence"].(float64); !ok {
		t.Fatalf("expected confidence field, got: %#v", payload["confidence"])
	}
	missingInfo, ok := payload["missing_information"].([]any)
	if !ok || len(missingInfo) == 0 {
		t.Fatalf("expected missing_information to be populated, got: %#v", payload["missing_information"])
	}
}

func TestEnrichAnalysisResult_PreservesExistingConfidence(t *testing.T) {
	raw := json.RawMessage(`{"summary":"ok","parties":[],"contract_type":"servicos","key_clauses":[],"risks":[],"recommendations":[],"overall_risk":"low","confidence":0.82}`)
	extractResult := &providers.ExtractResult{
		Text:         "texto",
		Source:       "native",
		QualityScore: 0.5,
	}

	enriched, err := enrichAnalysisResult(raw, extractResult)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	payload := map[string]any{}
	if err := json.Unmarshal(enriched, &payload); err != nil {
		t.Fatalf("unexpected unmarshal error: %v", err)
	}

	confidence, ok := payload["confidence"].(float64)
	if !ok {
		t.Fatalf("expected confidence float64, got: %#v", payload["confidence"])
	}
	if confidence != 0.82 {
		t.Fatalf("expected existing confidence to be preserved, got: %.2f", confidence)
	}
}
