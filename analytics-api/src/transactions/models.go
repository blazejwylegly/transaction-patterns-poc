package transactions

import (
	"github.com/google/uuid"
	"time"
)

type TxnContext struct {
	TxnId        uuid.UUID
	TxnName      string
	TxnStartedAt time.Time
	TxnStatus    string
}

type TxnCompleted struct {
	OrderID    uuid.UUID `json:"order_id"`
	CustomerID uuid.UUID `json:"customer_id"`
	Details    string    `json:"details"`
	Status     string    `json:"status"`
}
