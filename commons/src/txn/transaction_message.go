package txn

import (
	"github.com/google/uuid"
	"time"
)

type TransactionMessage[PayloadType any] struct {
	Id           uuid.UUID
	StepName     string
	StepExecutor string
	Name         string
	Payload      PayloadType
	InitiatedAt  time.Time
}
