package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/hahahamid/broker-backend/internal/models"
)

type OrderbookHandler struct{}

func NewOrderbookHandler() *OrderbookHandler {
	return &OrderbookHandler{}
}

func (h *OrderbookHandler) Get(c *gin.Context) {
	data := []models.Order{
		{ID: "1", Symbol: "AAPL", Side: "buy", Quantity: 10, Price: 150, RealizedPNL: 0, UnrealizedPNL: 20},
		{ID: "2", Symbol: "TSLA", Side: "sell", Quantity: 2, Price: 700, RealizedPNL: 50, UnrealizedPNL: 0},
	}
	c.JSON(http.StatusOK, gin.H{
		"orders": data,
		"card":   gin.H{"realized_pnl": 50, "unrealized_pnl": 20},
	})
}
