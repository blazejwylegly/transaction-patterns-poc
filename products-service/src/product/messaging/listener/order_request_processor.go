package listener

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/events"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/saga"
)

type OrderRequestProcessor struct {
	coordinator saga.Coordinator
}

func NewOrderRequestProcessor(coordinator saga.Coordinator) *OrderRequestProcessor {
	return &OrderRequestProcessor{coordinator: coordinator}
}

func (processor *OrderRequestProcessor) process(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		txMessage, err := parseEvent(msg)
		headers := parseHeaders(msg.Headers)
		processor.coordinator.HandleTransaction(*txMessage, headers)
		if err != nil {
			fmt.Printf("Error parsing msg { partition:'%d', offset:'%d' }: %v\n", msg.Partition, msg.Offset, err)
		}
	}
}

func parseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}

func parseEvent(msg *sarama.ConsumerMessage) (*events.OrderPlaced, error) {
	orderPlacedEvent := events.OrderPlaced{}
	err := json.Unmarshal(msg.Value, &orderPlacedEvent)
	if err != nil {
		fmt.Printf("Error parsing transaction message: %v\n", err)
		return nil, err
	}
	return &orderPlacedEvent, nil
}
