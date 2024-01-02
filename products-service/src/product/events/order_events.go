package events

import (
	"github.com/google/uuid"
	"time"
)

type OrderItemsReserved struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	TotalCost  float64   `json:"total_cost"`
}

type OrderPlaced struct {
	OrderID    uuid.UUID   `json:"order_id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	OrderItems []OrderItem `json:"order_items"`
	PlacedAt   time.Time   `json:"placed_at"`
}

type OrderItem struct {
	ProductId       uuid.UUID `json:"product_id"`
	QuantityOrdered int       `json:"quantity_ordered"`
}

type OrderFailed struct {
	OrderID uuid.UUID `json:"order_id"`
}
