package usecase

import (
	"context"
	"fmt"
	"time"

	"contract-scanner/internal/infra/database/postgres/models"
	"contract-scanner/internal/usecase/providers"
	"contract-scanner/internal/usecase/repositories"

	"github.com/google/uuid"
)

type IGeneratePresignedUrl interface {
	Execute(ctx context.Context, input PresignInput) (*Output, error)
}

type PresignInput struct {
	Filename    string
	ContentType string
	SizeBytes   int64
	ClerkUserID string
}

type Output struct {
	AnalysisID string `json:"analysis_id"`
	UploadURL  string `json:"upload_url"`
	S3Key      string `json:"s3_key"`
}

type GeneratePresignedUrl struct {
	AnalyseRepo     repositories.AnalyseRepo
	StorageProvider providers.StorageProvider
}

func NewGeneratePresignedUrl(analyseRepo repositories.AnalyseRepo, storageProvider providers.StorageProvider) IGeneratePresignedUrl {
	return &GeneratePresignedUrl{
		AnalyseRepo:     analyseRepo,
		StorageProvider: storageProvider,
	}
}

func (uc *GeneratePresignedUrl) Execute(ctx context.Context, input PresignInput) (*Output, error) {
	analyseID := uuid.New()

	s3Key := fmt.Sprintf("tmp/%s.pdf", analyseID)

	analyse := models.Analyse{
		ID:          analyseID,
		ClerkUserID: input.ClerkUserID,
		Status:      "UPLOADED",
		Filename:    &input.Filename,
		ContentType: &input.ContentType,
		SizeBytes:   &input.SizeBytes,
		S3Key:       s3Key,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	if err := uc.AnalyseRepo.Create(analyse); err != nil {
		return nil, fmt.Errorf("error creating analyse: %w", err)
	}

	uploadURL, err := uc.StorageProvider.GeneratePresignedPutURL(ctx, s3Key, input.ContentType)
	if err != nil {
		return nil, fmt.Errorf("error generating presigned url: %w", err)
	}

	return &Output{
		AnalysisID: analyseID.String(),
		UploadURL:  uploadURL,
		S3Key:      s3Key,
	}, nil
}
