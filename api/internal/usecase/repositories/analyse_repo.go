package repositories

import (
	"contract-scanner/internal/infra/database/postgres/models"

	"github.com/google/uuid"
)

type AnalyseRepo interface {
	Create(analyse models.Analyse) error
	Get(id uuid.UUID) (*models.Analyse, error)
	Update(analyse *models.Analyse) error
}
