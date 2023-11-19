package orderModels

import (
	"github.com/google/uuid"
	"time"
)

type OrderRequestDto struct {
	OrderID    uuid.UUID   `json:"order_id"`
	CustomerID uuid.UUID   `json:"customer_id"`
	OrderItems []OrderItem `json:"order_items"`
	PlacedAt   time.Time   `json:"placed_at"`
}

type OrderItem struct {
	ProductId       uuid.UUID `json:"product_id"`
	QuantityOrdered int       `json:"quantity_ordered"`
}
