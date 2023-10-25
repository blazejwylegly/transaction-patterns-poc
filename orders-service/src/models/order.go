package models

import (
	"github.com/google/uuid"
	"time"
)

type Order struct {
	OrderID    uuid.UUID   `json:"order_id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	OrderItems []OrderItem `json:"order_items"`
	PlacedAt   time.Time   `json:"placed_at"`
}

func NewOrder() *Order {
	return &Order{
		OrderID:    uuid.New(),
		OrderItems: make([]OrderItem, 0),
		PlacedAt:   time.Now(),
	}
}
