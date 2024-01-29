package database

import (
	"github.com/google/uuid"
	"time"
)

type Transaction struct {
	TxnId     uuid.UUID `gorm:"type:uuid;column:txn_id;primary_key"`
	TxnName   string
	StartedAt time.Time
	Steps     []TransactionStep `gorm:"foreignKey:TxnId"`
	Status    string
}

type TransactionStep struct {
	StepId       uuid.UUID `gorm:"type:uuid;column:step_id;primary_key"`
	TxnId        uuid.UUID
	StepName     string
	StepExecutor string
	StepStatus   string
	Payload      string `gorm:"type:jsonb" json:"payload"`
	Headers      string `gorm:"type:jsonb" json:"headers"`
}
