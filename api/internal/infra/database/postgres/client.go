package postgres

import (
	"fmt"

	"contract-scanner/internal/infra/database/postgres/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type DatabaseConfig struct {
	Host     string
	Name     string
	User     string
	Password string
	Port     string
	SslMode  string
}

type Client struct {
	dataSourceName string
}

func NewClient(config DatabaseConfig) *Client {
	url_connection := fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s port=%s sslmode=%s",
		config.Host,
		config.User,
		config.Password,
		config.Name,
		config.Port,
		config.SslMode,
	)
	return &Client{
		dataSourceName: url_connection,
	}
}

func (c *Client) Open() (*gorm.DB, error) {
	db, err := gorm.Open(postgres.Open(c.dataSourceName), &gorm.Config{})
	if err != nil {
		fmt.Println("error opening database connection: ", err)
		return nil, err
	}

	return db, nil
}

func (c *Client) Migrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&models.Analyse{},
	)
}
