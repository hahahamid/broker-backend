package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hahahamid/broker-backend/internal/models"
)

type HoldingsHandler struct{}

func NewHoldingsHandler() *HoldingsHandler {
	return &HoldingsHandler{}
}

func (h *HoldingsHandler) Get(c *gin.Context) {
	data := []models.Holding{
		{Symbol: "AAPL", Quantity: 10, AvgPrice: 150},
		{Symbol: "GOOGL", Quantity: 5, AvgPrice: 2500},
	}
	c.JSON(http.StatusOK, data)
}
