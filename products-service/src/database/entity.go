package database

import "github.com/google/uuid"

type Product struct {
	ProductID   uuid.UUID `gorm:"type:uuid;column:product_id;primary_key"`
	Name        string    `gorm:"column:name"`
	Price       float64   `gorm:"column:price"`
	Description string    `gorm:"column:description"`
	Quantity    int       `gorm:"column:quantity"`
}

type TxLogItem struct {
	TxLogItemId uuid.UUID `gorm:"type:uuid;primary_key"`
	OrderID     uuid.UUID ` gorm:"type:uuid;"`
	OrderItems  string    `gorm:"type:jsonb"`
}
