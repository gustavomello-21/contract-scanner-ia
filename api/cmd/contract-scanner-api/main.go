package main

import (
	"log"
	"os"

	"contract-scanner/cmd/contract-scanner-api/config"
	"contract-scanner/internal/handler"
	postgres "contract-scanner/internal/infra/database/postgres"
	"contract-scanner/internal/infra/database/postgres/repository"
	"contract-scanner/internal/infra/llm"
	"contract-scanner/internal/infra/pdf"
	"contract-scanner/internal/infra/storage"
	"contract-scanner/internal/usecase"

	"github.com/clerk/clerk-sdk-go/v2"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	_ = godotenv.Load("../.env")
	_ = godotenv.Load("../../../.env")

	clerk.SetKey(os.Getenv("CLERK_SECRET_KEY"))

	dbClient := postgres.NewClient(postgres.DatabaseConfig{
		Host:     os.Getenv("DB_HOST"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Port:     os.Getenv("DB_PORT"),
		SslMode:  os.Getenv("DB_SSLMODE"),
	})

	db, err := dbClient.Open()
	if err != nil {
		log.Fatal("error connecting to database: ", err)
	}

	if err := dbClient.Migrate(db); err != nil {
		log.Fatal("error running migrations: ", err)
	}

	log.Println("database connected and migrated")

	s3Client := storage.NewS3Client(
		os.Getenv("AWS_BUCKET"),
		os.Getenv("AWS_REGION"),
		os.Getenv("ACCESS_KEY"),
		os.Getenv("SECRET_ACCESS_KEY"),
	)

	analyseRepo := repository.NewAnalyseRepo(db)

	pdfExtractor := pdf.NewPdfCpuExtractor()
	openaiClient := llm.NewOpenAIClient(os.Getenv("OPEN_AI_API_KEY"))

	generatePresignedUrl := usecase.NewGeneratePresignedUrl(analyseRepo, s3Client)
	processContract := usecase.NewProcessContract(analyseRepo, s3Client, pdfExtractor, openaiClient)

	uploadHandler := handler.NewUploadHandler(generatePresignedUrl)
	analyseHandler := handler.NewAnalyseHandler(processContract)

	r := config.Routes(uploadHandler, analyseHandler)

	log.Println("server running on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatal(err)
	}
}
