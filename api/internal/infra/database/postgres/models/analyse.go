package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/datatypes"
)

type Analyse struct {
	ID            uuid.UUID      `gorm:"type:uuid;primaryKey"`
	ClerkUserID   string         `gorm:"type:text;not null"`
	Status        string         `gorm:"type:text;not null"`
	Filename      *string        `gorm:"type:text"`
	ContentType   *string        `gorm:"type:text"`
	SizeBytes     *int64         `gorm:"type:bigint"`
	S3Bucket      *string        `gorm:"type:text"`
	S3Key         string         `gorm:"type:text;not null"`
	Model         string         `gorm:"type:text;default:'gpt-4o-mini'"`
	PromptVersion string         `gorm:"type:text;default:'v1'"`
	CharCount     *int           `gorm:"type:int"`
	ResultJSON    datatypes.JSON `gorm:"type:jsonb"`
	CreatedAt     time.Time      `gorm:"autoCreateTime"`
	UpdatedAt     time.Time      `gorm:"autoUpdateTime"`
	CompletedAt   *time.Time     `gorm:"type:timestamptz"`
}
