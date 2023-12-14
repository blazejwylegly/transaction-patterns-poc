package transactions

import (
	"github.com/google/uuid"
	"time"
)

type TxnContext struct {
	TxnId        uuid.UUID
	TxnName      string
	TxnStartedAt time.Time
}
