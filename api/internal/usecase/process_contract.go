package usecase

import (
	"context"
	"encoding/json"
	"fmt"
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
	text, err := uc.PDFExtractor.Extract(tmpPath)
	if err != nil {
		return nil, fmt.Errorf("error extracting text: %w", err)
	}
	fmt.Printf(text)
	// 7. Chama LLM
	resultJSON, err := uc.LLMProvider.AnalyzeContract(ctx, text)
	if err != nil {
		return nil, fmt.Errorf("error analyzing contract: %w", err)
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
