package handler

import (
	"net/http"
	"os"

	"contract-scanner/internal/usecase"

	"github.com/gin-gonic/gin"
)

type UploadHandler struct {
	generatePresignedUrl usecase.IGeneratePresignedUrl
}

func NewUploadHandler(generatePresignedUrl usecase.IGeneratePresignedUrl) *UploadHandler {
	return &UploadHandler{generatePresignedUrl: generatePresignedUrl}
}

type PresignRequest struct {
	Filename    string `json:"filename" binding:"required"`
	ContentType string `json:"content_type" binding:"required"`
	SizeBytes   int64  `json:"size_bytes" binding:"required,gt=0"`
}

func (h *UploadHandler) Presign(c *gin.Context) {
	var req PresignRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	output, err := h.generatePresignedUrl.Execute(c.Request.Context(), usecase.PresignInput{
		Filename:    req.Filename,
		ContentType: req.ContentType,
		SizeBytes:   req.SizeBytes,
		ClerkUserID: "", // disabled for local testing
		Model:       os.Getenv("LLM_MODEL"),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, output)
}
