package application

import (
	"errors"
	"fmt"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/database"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type PaymentRequestedHandler struct {
	db *gorm.DB
}

func NewPaymentRequestedHandler(db *gorm.DB) *PaymentRequestedHandler {
	return &PaymentRequestedHandler{db: db}
}

func (handler *PaymentRequestedHandler) Handle(inputEvent events.PaymentRequested) (*events.PaymentProcessed, error) {
	var outputEvent *events.PaymentProcessed
	err := handler.db.Transaction(func(tx *gorm.DB) error {
		// Retrieve user
		var customer *database.Customer
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Preload("Payments").
			First(&database.Customer{}).
			Error
		if err != nil {
			log.Printf("Customer with id %s not found\n", inputEvent.CustomerID)
			return err
		}
		if customer == nil {
			log.Printf("Customer with id %s not found\n", inputEvent.CustomerID)
			return errors.New(fmt.Sprintf("Customer with id %s not found", inputEvent.CustomerID))
		}

		// Check user funds
		if customer.AvailableFunds < inputEvent.TotalCost {
			log.Printf("Customer %s has not enough funds to pay for orderId %s\n",
				inputEvent.CustomerID, inputEvent.OrderID)
			return errors.New(fmt.Sprintf("Customer with id %s not found", inputEvent.CustomerID))
		}

		// Lower his funds
		err = tx.Model(customer).Update("available_funds", customer.AvailableFunds-inputEvent.TotalCost).Error
		if err != nil {
			log.Printf("Error trying to decrease customer funds for orderId %s: %v\n", inputEvent.OrderID, err)
			return err
		}

		// Register payment
		payment := &database.Payment{
			PaymentId:  uuid.New(),
			OrderId:    inputEvent.OrderID,
			CustomerId: customer.CustomerId,
			TotalCost:  inputEvent.TotalCost,
			PaidWith:   events.CARD,
		}
		err = tx.Save(payment).Error

		if err != nil {
			log.Printf("Error saving payment for orderId %s: %v\n", inputEvent.OrderID, err)
			return err
		}

		//  Create domain inputEvent
		outputEvent = &events.PaymentProcessed{
			OrderID:    payment.OrderId,
			CustomerID: payment.CustomerId,
			PaidWith:   payment.PaidWith,
		}
		return nil
	})

	if err != nil {
		log.Printf("Error processing event for orderId %s: %v\n", inputEvent.OrderID, err)
		return nil, err
	}
	return outputEvent, nil
}
