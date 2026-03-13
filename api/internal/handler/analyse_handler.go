package handler

import (
	"log"
	"net/http"

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

	log.Printf("[process] starting analyse_id=%s", id)

	output, err := h.processContract.Execute(c.Request.Context(), usecase.ProcessInput{
		AnalyseID:   id,
		ClerkUserID: "", // disabled for local testing
	})
	if err != nil {
		log.Printf("[process] failed analyse_id=%s error=%v", id, err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	log.Printf("[process] completed analyse_id=%s status=%s", id, output.Status)
	c.JSON(http.StatusOK, output)
}
