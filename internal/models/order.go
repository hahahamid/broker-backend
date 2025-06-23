package models

type Order struct {
	ID            string  `json:"id"`
	Symbol        string  `json:"symbol"`
	Side          string  `json:"side"` // "buy" or "sell"
	Quantity      float64 `json:"quantity"`
	Price         float64 `json:"price"`
	RealizedPNL   float64 `json:"realized_pnl"`
	UnrealizedPNL float64 `json:"unrealized_pnl"`
}
