package application

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"gorm.io/gorm"
)

type PaymentRequestedHandler struct {
	db *gorm.DB
}

func NewItemsReservedEventHandler(db *gorm.DB) *PaymentRequestedHandler {
	return &PaymentRequestedHandler{db}
}

func (handler *PaymentRequestedHandler) Handle(event events.PaymentRequested) (*events.PaymentProcessed, error) {
	// TODO
	return nil, nil
}
