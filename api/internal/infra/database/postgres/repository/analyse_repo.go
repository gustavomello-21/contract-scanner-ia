package repository

import (
	"contract-scanner/internal/infra/database/postgres/models"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type AnalyseRepoPostgres struct {
	db *gorm.DB
}

func NewAnalyseRepo(db *gorm.DB) *AnalyseRepoPostgres {
	return &AnalyseRepoPostgres{db: db}
}

func (r *AnalyseRepoPostgres) Create(analyse models.Analyse) error {
	return r.db.Create(&analyse).Error
}

func (r *AnalyseRepoPostgres) Get(id uuid.UUID) (*models.Analyse, error) {
	var analyse models.Analyse
	if err := r.db.Where("id = ?", id).First(&analyse).Error; err != nil {
		return nil, err
	}
	return &analyse, nil
}

func (r *AnalyseRepoPostgres) Update(analyse *models.Analyse) error {
	return r.db.Save(analyse).Error
}
