package database

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"github.com/google/uuid"
)

type Customer struct {
	CustomerId     uuid.UUID `gorm:"type:uuid;column:customer_id;primary_key"`
	AvailableFunds float64
	Payments       []Payment
}

type Payment struct {
	PaymentId  uuid.UUID `gorm:"type:uuid;column:payment_id;primary_key"`
	OrderId    uuid.UUID
	CustomerId uuid.UUID
	TotalCost  float64
	PaidWith   events.PaymentType `gorm:"type:string;column:payment_type"`
}
