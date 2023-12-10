package db

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	TxnId     uuid.UUID `gorm:"type:uuid;column:txn_id;primary_key"`
	TxnName   string
	StartedAt time.Time
	Steps     []TransactionStep `gorm:"foreignKey:StepId"`
}

type TransactionStep struct {
	StepId       uuid.UUID `gorm:"type:uuid;column:txn_id;primary_key"`
	StepName     string
	StepExecutor string
	Payload      string `gorm:"type:jsonb" json:"payload"`
	Headers      string `gorm:"type:jsonb" json:"headers"`
}
