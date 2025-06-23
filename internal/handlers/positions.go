package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hahahamid/broker-backend/internal/models"
)

type PositionsHandler struct{}

func NewPositionsHandler() *PositionsHandler {
	return &PositionsHandler{}
}

func (h *PositionsHandler) Get(c *gin.Context) {
	data := []models.Position{
		{Symbol: "AAPL", Quantity: 10, AvgPrice: 150, PNL: 20},
		{Symbol: "TSLA", Quantity: 2, AvgPrice: 700, PNL: 50},
	}
	c.JSON(http.StatusOK, gin.H{"positions": data})
}
