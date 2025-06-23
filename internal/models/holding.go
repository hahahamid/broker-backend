package models

type Holding struct {
	Symbol   string  `json:"symbol"`
	Quantity float64 `json:"quantity"`
	AvgPrice float64 `json:"avg_price"`
}
