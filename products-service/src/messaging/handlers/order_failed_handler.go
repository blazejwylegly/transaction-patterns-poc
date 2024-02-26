package handlers

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application/saga"
)

type OrderFailedHandler struct {
	coordinator saga.Coordinator
}

func NewOrderFailedHandler(coordinator saga.Coordinator) *OrderFailedHandler {
	return &OrderFailedHandler{coordinator: coordinator}
}

func (handler *OrderFailedHandler) Handle(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		event := &application.OrderFailed{}
		err := parseEvent(event, msg)
		if err != nil {
			fmt.Printf("Error parsing msg { partition:'%d', offset:'%d' }: %v\n", msg.Partition, msg.Offset, err)
		}

		headers := parseHeaders(msg.Headers)
		handler.coordinator.HandleSagaRollbackEvent(*event, headers)
	}
}
