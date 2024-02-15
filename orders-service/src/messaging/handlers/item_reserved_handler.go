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

type ItemReservationStatusHandler struct {
	sagaCoordinator saga.OrchestrationCoordinator
}

func NewItemReservationStatusHandler(coordinator saga.OrchestrationCoordinator) *ItemReservationStatusHandler {
	return &ItemReservationStatusHandler{sagaCoordinator: coordinator}
}

func (handler *ItemReservationStatusHandler) Handle() func(chan *sarama.ConsumerMessage) {
	return func(messagesChannel chan *sarama.ConsumerMessage) {
		for msg := range messagesChannel {
			itemReservationStatus, err := parseItemReservationStatus(msg)
			if err != nil {
				log.Printf("Error trying to parse event for message with offset %d", msg.Offset)
			}
			headers := messaging.ParseHeaders(msg.Headers)
			context, err := saga.ContextFromHeaders(headers)
			handler.sagaCoordinator.HandleItemReservationStatusEvent(*itemReservationStatus,
				*context)
		}
	}

}

func parseItemReservationStatus(msg *sarama.ConsumerMessage) (*application.ItemReservationStatus, error) {
	itemReservationStatus := application.ItemReservationStatus{}
	err := json.Unmarshal(msg.Value, &itemReservationStatus)
	if err != nil {
		fmt.Printf("Error parsing transaction message: %v\n", err)
		return nil, err
	}
	return &itemReservationStatus, nil
}
