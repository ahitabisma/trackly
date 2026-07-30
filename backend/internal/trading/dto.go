package trading

import (
	"time"

	"trackly-backend/internal/analisis"
)

type TransactionRequest struct {
	Ticker          string  `json:"ticker" validate:"required"`
	TransactionType string  `json:"transaction_type" validate:"required,oneof=buy sell"`
	Lot             float64 `json:"lot" validate:"required,gt=0"`
	Price           float64 `json:"price" validate:"required,gt=0"`
	TransactionDate string  `json:"transaction_date" validate:"required"`
	Notes           string  `json:"notes,omitempty"`
}

type Transaction struct {
	ID              string    `json:"id"`
	UserID          uint      `json:"user_id"`
	Ticker          string    `json:"ticker"`
	TransactionType string    `json:"transaction_type"`
	Lot             float64   `json:"lot"`
	Price           float64   `json:"price"`
	TransactionDate string    `json:"transaction_date"`
	Notes           *string   `json:"notes,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type UpdateTransactionRequest struct {
	Ticker          *string  `json:"ticker,omitempty"`
	TransactionType *string  `json:"transaction_type,omitempty"`
	Lot             *float64 `json:"lot,omitempty"`
	Price           *float64 `json:"price,omitempty"`
	TransactionDate *string  `json:"transaction_date,omitempty"`
	Notes           *string  `json:"notes,omitempty"`
}

type Position struct {
	Ticker          string  `json:"ticker"`
	TotalLot        float64 `json:"total_lot"`
	AverageBuyPrice float64 `json:"average_buy_price"`
	Status          string  `json:"status"`
}

type PositionReviewResponse struct {
	Ticker           string                 `json:"ticker"`
	Position         Position               `json:"position"`
	Indicators       interface{}            `json:"indicators,omitempty"`
	Signal           interface{}            `json:"signal,omitempty"`
	PositionReview   map[string]interface{} `json:"position_review"`
	Snapshot         *analisis.Snapshot     `json:"snapshot,omitempty"`
	AIInsight        string                 `json:"ai_insight,omitempty"`
}
