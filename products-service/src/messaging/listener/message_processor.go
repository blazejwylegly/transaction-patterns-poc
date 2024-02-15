package listener

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application/saga"
)

type OrderRequestProcessor struct {
	coordinator saga.Coordinator
}

type OrderFailedProcessor struct {
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

func NewOrderFailedProcessor(coordinator saga.Coordinator) *OrderFailedProcessor {
	return &OrderFailedProcessor{coordinator: coordinator}
}

func (processor *OrderFailedProcessor) process(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		event := &application.OrderFailed{}
		err := parseEvent(event, msg)
		if err != nil {
			fmt.Printf("Error parsing msg { partition:'%d', offset:'%d' }: %v\n", msg.Partition, msg.Offset, err)
		}

		headers := parseHeaders(msg.Headers)
		processor.coordinator.HandleRollback(*event, headers)
	}
}

func NewOrderRequestProcessor(coordinator saga.Coordinator) *OrderRequestProcessor {
	return &OrderRequestProcessor{coordinator: coordinator}
}

func (processor *OrderRequestProcessor) process(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		event := &application.OrderPlaced{}
		err := parseEvent(event, msg)
		if err != nil {
			fmt.Printf("Error parsing msg { partition:'%d', offset:'%d' }: %v\n", msg.Partition, msg.Offset, err)
		}

		headers := parseHeaders(msg.Headers)
		processor.coordinator.HandleTransaction(*event, headers)
	}
}
