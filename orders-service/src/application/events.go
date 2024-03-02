package application

import (
	"github.com/google/uuid"
	"time"
)

type OrderPlaced struct {
	OrderID    uuid.UUID   `json:"order_id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	OrderItems []OrderItem `json:"order_items"`
	PlacedAt   time.Time   `json:"placed_at"`
}

type OrderFailed struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Details    string    `json:"details"`
}

type OrderCompleted struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Details    string    `json:"details"`
}

type ItemReservationStatus struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	TotalCost  float64   `json:"total_cost"`
	Status     string    `json:"status"`
}

type PaymentRequest struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	TotalCost  float64   `json:"total_cost"`
}

type ItemReservationRollbackRequest struct {
	OrderID    uuid.UUID `json:"order_id"`
	IsRollback bool      `json:"is_rollback"`
}

type PaymentStatus struct {
	OrderID    uuid.UUID   `json:"order_id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	Status     string      `json:"payment_status"`
	PaidWith   PaymentType `json:"paid_with"`
}

type OrderStatus struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Status     string    `json:"status"`
}
