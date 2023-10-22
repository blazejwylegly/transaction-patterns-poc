package models

type ProductDto struct {
	ProductId   int     `json:"productId"`
	Name        string  `json:"name"`
	Price       float64 `json:"price"`
	Description string  `json:"description"`
	Quantity    int     `json:"quantity"`
}
