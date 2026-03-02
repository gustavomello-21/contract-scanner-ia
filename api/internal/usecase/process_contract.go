package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"contract-scanner/internal/usecase/providers"
	"contract-scanner/internal/usecase/repositories"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type IProcessContract interface {
	Execute(ctx context.Context, input ProcessInput) (*ProcessOutput, error)
}

type ProcessInput struct {
	AnalyseID   uuid.UUID
	ClerkUserID string
}

type ProcessOutput struct {
	AnalysisID string          `json:"analysis_id"`
	Status     string          `json:"status"`
	Result     json.RawMessage `json:"result"`
}

type ProcessContract struct {
	AnalyseRepo     repositories.AnalyseRepo
	StorageProvider providers.StorageProvider
	PDFExtractor    providers.PDFExtractor
	LLMProvider     providers.LLMProvider
}

func NewProcessContract(
	analyseRepo repositories.AnalyseRepo,
	storageProvider providers.StorageProvider,
	pdfExtractor providers.PDFExtractor,
	llmProvider providers.LLMProvider,
) IProcessContract {
	return &ProcessContract{
		AnalyseRepo:     analyseRepo,
		StorageProvider: storageProvider,
		PDFExtractor:    pdfExtractor,
		LLMProvider:     llmProvider,
	}
}

func (uc *ProcessContract) Execute(ctx context.Context, input ProcessInput) (*ProcessOutput, error) {
	// 1. Busca analyse
	analyse, err := uc.AnalyseRepo.Get(input.AnalyseID)
	if err != nil {
		return nil, fmt.Errorf("analyse not found: %w", err)
	}

	// 2. Valida dono
	if analyse.ClerkUserID != input.ClerkUserID {
		return nil, fmt.Errorf("forbidden: user does not own this analyse")
	}

	if analyse.Status == "PROCESSING" {
		return nil, fmt.Errorf("File already processing")
	}

	// 3. Verifica se arquivo existe no S3
	if err := uc.StorageProvider.HeadObject(ctx, analyse.S3Key); err != nil {
		return nil, fmt.Errorf("file not uploaded yet: %w", err)
	}

	// 4. Update status = PROCESSING
	analyse.Status = "PROCESSING"
	if err := uc.AnalyseRepo.Update(analyse); err != nil {
		return nil, fmt.Errorf("error updating status: %w", err)
	}

	// 5. Download do S3 para /tmp
	tmpPath := fmt.Sprintf("/tmp/%s.pdf", analyse.ID.String())
	if err := uc.StorageProvider.GetObject(ctx, analyse.S3Key, tmpPath); err != nil {
		return nil, fmt.Errorf("error downloading file: %w", err)
	}

	// 6. Extrai texto do PDF
	extractResult, err := uc.PDFExtractor.Extract(ctx, tmpPath)
	if err != nil {
		return nil, fmt.Errorf("error extracting text: %w", err)
	}
	text := extractResult.Text
	if strings.TrimSpace(text) == "" {
		return nil, fmt.Errorf("unable to extract readable contract text")
	}

	workDir, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting workdir: %w", err)
	}

	tmpDir := filepath.Join(workDir, "tmp")
	if err := os.MkdirAll(tmpDir, 0o755); err != nil {
		return nil, fmt.Errorf("error creating tmp dir: %w", err)
	}

	tmpTextFilename := buildTempTextFilename(analyse.ID.String(), analyse.Filename)
	tmpTextPath := filepath.Join(tmpDir, tmpTextFilename)

	if err := os.WriteFile(tmpTextPath, []byte(text), 0o644); err != nil {
		return nil, fmt.Errorf("error writing extracted text to temporary file: %w", err)
	}

	log.Printf(
		"texto extraido salvo temporariamente em: %s (source=%s quality=%.2f warnings=%v)",
		tmpTextPath,
		extractResult.Source,
		extractResult.QualityScore,
		extractResult.Warnings,
	)

	// 7. Chama LLM
	resultJSON, err := uc.LLMProvider.AnalyzeContract(ctx, text)
	if err != nil {
		analyse.Status = "FAILED"
		if err := uc.AnalyseRepo.Update(analyse); err != nil {
			return nil, fmt.Errorf("error updating status: %w", err)
		}
		return nil, fmt.Errorf("error analyzing contract: %w", err)
	}
	resultJSON, err = enrichAnalysisResult(resultJSON, extractResult)
	if err != nil {
		return nil, fmt.Errorf("error enriching analysis result: %w", err)
	}

	// 8. Salva resultado e status = COMPLETED
	analyse.ResultJSON = datatypes.JSON(resultJSON)
	analyse.Status = "COMPLETED"
	now := time.Now().UTC()
	analyse.CompletedAt = &now
	if err := uc.AnalyseRepo.Update(analyse); err != nil {
		return nil, fmt.Errorf("error saving result: %w", err)
	}

	// 9. Retorna output
	return &ProcessOutput{
		AnalysisID: analyse.ID.String(),
		Status:     analyse.Status,
		Result:     resultJSON,
	}, nil
}

func buildTempTextFilename(fallbackID string, originalFilename *string) string {
	if originalFilename == nil || strings.TrimSpace(*originalFilename) == "" {
		return fmt.Sprintf("%s.txt", fallbackID)
	}

	base := filepath.Base(strings.TrimSpace(*originalFilename))
	name := strings.TrimSuffix(base, filepath.Ext(base))
	name = strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z':
			return r
		case r >= 'A' && r <= 'Z':
			return r
		case r >= '0' && r <= '9':
			return r
		case r == '-', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
	name = strings.Trim(name, "_")
	if name == "" {
		name = fallbackID
	}

	return fmt.Sprintf("%s.txt", name)
}

func enrichAnalysisResult(raw json.RawMessage, extractResult *providers.ExtractResult) (json.RawMessage, error) {
	payload := map[string]any{}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return nil, fmt.Errorf("invalid result json: %w", err)
	}

	warnings := []string{}
	if current, ok := payload["analysis_warnings"]; ok {
		switch values := current.(type) {
		case []any:
			for _, value := range values {
				if warning, ok := value.(string); ok {
					trimmed := strings.TrimSpace(warning)
					if trimmed != "" && !slices.Contains(warnings, trimmed) {
						warnings = append(warnings, trimmed)
					}
				}
			}
		case []string:
			for _, warning := range values {
				trimmed := strings.TrimSpace(warning)
				if trimmed != "" && !slices.Contains(warnings, trimmed) {
					warnings = append(warnings, trimmed)
				}
			}
		}
	}

	for _, warning := range extractResult.Warnings {
		trimmed := strings.TrimSpace(warning)
		if trimmed != "" && !slices.Contains(warnings, trimmed) {
			warnings = append(warnings, trimmed)
		}
	}
	if extractResult.QualityScore < 0.35 && !slices.Contains(warnings, "text extraction quality is low; analysis may be partial") {
		warnings = append(warnings, "text extraction quality is low; analysis may be partial")
	}
	if len(warnings) > 0 {
		payload["analysis_warnings"] = warnings
	}

	if _, exists := payload["confidence"]; !exists {
		payload["confidence"] = math.Round(extractResult.QualityScore*100) / 100
	}

	if _, exists := payload["missing_information"]; !exists && extractResult.QualityScore < 0.35 {
		payload["missing_information"] = []string{
			"O texto extraído possui baixa qualidade e pode omitir cláusulas relevantes.",
		}
	}

	out, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}
	return json.RawMessage(out), nil
}
