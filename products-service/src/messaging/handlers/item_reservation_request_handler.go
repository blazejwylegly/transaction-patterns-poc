package handlers

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application/saga"
)

type ItemReservationRequestHandler struct {
	coordinator saga.Coordinator
}

func parseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}

func parseEvent[T interface{}](event *T, msg *sarama.ConsumerMessage) error {
	err := json.Unmarshal(msg.Value, &event)
	if err != nil {
		fmt.Printf("Error parsing transaction message: %v\n", err)
		return err
	}
	return nil
}

func NewItemReservationRequestHandler(coordinator saga.Coordinator) *ItemReservationRequestHandler {
	return &ItemReservationRequestHandler{coordinator: coordinator}
}

func (handler *ItemReservationRequestHandler) Handle(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		event := &application.InventoryUpdateRequest{}
		err := parseEvent(event, msg)
		if err != nil {
			fmt.Printf("Error parsing msg { partition:'%d', offset:'%d' }: %v\n", msg.Partition, msg.Offset, err)
		}

		headers := parseHeaders(msg.Headers)
		handler.coordinator.HandleSagaEvent(*event, headers)
	}
}
