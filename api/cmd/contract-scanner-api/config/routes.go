package config

import (
	"net/http"

	"contract-scanner/internal/handler"
	"contract-scanner/internal/middleware"

	"github.com/gin-gonic/gin"
)

func Routes(uploadHandler *handler.UploadHandler, analyseHandler *handler.AnalyseHandler) *gin.Engine {
	r := gin.Default()

	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	api := r.Group("/api")
	api.Use(middleware.ClerkAuth())
	{
		uploads := api.Group("/uploads")
		{
			uploads.POST("/presign", uploadHandler.Presign)
		}

		analyses := api.Group("/analyses")
		{
			analyses.POST("/:id/process", analyseHandler.Process)
		}
	}

	return r
}
