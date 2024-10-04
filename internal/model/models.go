package model

import (
	"errors"
	"time"
)

type User struct {
	ID          int       `json:"id"`
	Address     string    `json:"address"`
	TotalPoints float64   `json:"total_points"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type Token struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Symbol    string    `json:"symbol"`
	Decimals  int64     `json:"decimals"`
	CreatedAt time.Time `json:"created_at"`
}

type SwapHistory struct {
	ID              int       `json:"id"`
	Token           string    `json:"token"`
	Account         string    `json:"account"`
	TransactionHash string    `json:"transaction_hash"`
	UsdValue        float64   `json:"usd_value"`
	LastUpdated     time.Time `json:"last_updated"`
	CreatedAt       time.Time `json:"created_at"`
}

type PointsHistory struct {
	ID          int       `json:"id"`
	Token       string    `json:"token"`
	Account     string    `json:"account"`
	Points      float64   `json:"points"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

// other
type UserSwapPercentage struct {
	Account    string  `json:"account"`
	TotalUSD   float64 `json:"total_usd"`
	Percentage float64 `json:"percentage"`
}

// ErrUserNotFound is returned when a user cannot be found.
var (
	ErrUserNotFound  = errors.New("user not found")
	ErrTokenNotFound = errors.New("token not found")
)
