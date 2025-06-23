package models

type Position struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	AvgPrice float64 `json:"avg_price"`
	PNL      float64 `json:"pnl"`
}
