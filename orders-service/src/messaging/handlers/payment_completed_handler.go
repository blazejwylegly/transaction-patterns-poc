package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/application"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/application/saga"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"log"
)

type PaymentStatusHandler struct {
	sagaCoordinator saga.OrchestrationCoordinator
}

func NewPaymentStatusHandler(sagaCoordinator saga.OrchestrationCoordinator) *PaymentStatusHandler {
	return &PaymentStatusHandler{sagaCoordinator: sagaCoordinator}
}

func (handler *PaymentStatusHandler) Handle() func(chan *sarama.ConsumerMessage) {
	return func(messagesChannel chan *sarama.ConsumerMessage) {
		for msg := range messagesChannel {
			paymentStatusEvent, err := parsePaymentStatusEvent(msg)
			if err != nil {
				log.Printf("Error trying to parse event for message with offset %d", msg.Offset)
			}
			headers := messaging.ParseHeaders(msg.Headers)
			context, err := saga.ContextFromHeaders(headers)
			handler.sagaCoordinator.HandlePaymentStatusEvent(*paymentStatusEvent,
				*context)
		}
	}
}

func parsePaymentStatusEvent(msg *sarama.ConsumerMessage) (*application.PaymentStatus, error) {
	paymentStatus := application.PaymentStatus{}
	err := json.Unmarshal(msg.Value, &paymentStatus)
	if err != nil {
		fmt.Printf("Error parsing transaction message: %v\n", err)
		return nil, err
	}
	return &paymentStatus, nil
}
