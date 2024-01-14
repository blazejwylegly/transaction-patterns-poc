package database

import (
	"github.com/google/uuid"
)

type TxLogItem struct {
	TxLogItemId uuid.UUID `gorm:"type:uuid;primary_key"`
	OrderID     uuid.UUID ` gorm:"type:uuid;"`
	OrderItems  string    `gorm:"type:jsonb"`
}
