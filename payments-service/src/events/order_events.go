package events

import (
	"github.com/google/uuid"
)

type PaymentRequested struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	TotalCost  float64   `json:"total_cost"`
}

type PaymentReceived struct {
	OrderID    uuid.UUID   `json:"order_id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	PaidWith   PaymentType `json:"paid_with"`
}

type PaymentType string

const (
	CARD        PaymentType = "card"
	ON_DELIVERY PaymentType = "on_delivery"
)