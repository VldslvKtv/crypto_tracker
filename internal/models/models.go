package models

type Coin struct {
	Name      string  `json:"coin"`
	Price     float64 `json:"price"`
	Timestamp int64   `json:"timestamp"`
}

type CoinRequest struct {
	Coin string `json:"coin"`
}

type GetPriceRequest struct {
	Coin      string `json:"coin" validate:"required"`
	Timestamp string `json:"timestamp" validate:"required"`
}
