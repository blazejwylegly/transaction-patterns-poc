package db

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
	StepId        uuid.UUID `gorm:"type:uuid;column:step_id;primary_key"`
	Topic         string
	TxnId         uuid.UUID
	StepName      string
	StepExecutor  string
	StepStatus    string
	StepStartedAt string
}
