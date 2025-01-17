package models

type Coin struct {
	Name      string  `json:"coin"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}
