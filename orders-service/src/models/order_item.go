package models

import "github.com/google/uuid"

type OrderItem struct {
	ProductId       uuid.UUID `json:"product_id"`
	QuantityOrdered int       `json:"quantity_ordered"`
}
