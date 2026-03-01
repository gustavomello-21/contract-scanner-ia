package handler

import (
	"net/http"

	"contract-scanner/internal/middleware"
	"contract-scanner/internal/usecase"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type AnalyseHandler struct {
	processContract usecase.IProcessContract
}

func NewAnalyseHandler(processContract usecase.IProcessContract) *AnalyseHandler {
	return &AnalyseHandler{processContract: processContract}
}

func (h *AnalyseHandler) Process(c *gin.Context) {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid analyse id"})
		return
	}

	output, err := h.processContract.Execute(c.Request.Context(), usecase.ProcessInput{
		AnalyseID:   id,
		ClerkUserID: middleware.GetClerkUserID(c),
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, output)
}
