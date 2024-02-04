package dto

import "github.com/google/uuid"

type ProductDto struct {
	ProductId   uuid.UUID `json:"productId"`
	Name        string    `json:"name"`
	Price       float64   `json:"price"`
	Description string    `json:"description"`
	Quantity    int       `json:"quantity"`
}
