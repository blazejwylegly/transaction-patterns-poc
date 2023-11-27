package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/events"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging/producer"
)

type Coordinator struct {
	orderEventHandler application.OrderEventHandler
	producer          producer.EventProducer
	topics            config.KafkaTopics
}

func NewCoordinator(orderEventHandler application.OrderEventHandler,
	producer producer.EventProducer,
	config config.KafkaConfig) *Coordinator {
	return &Coordinator{
		orderEventHandler: orderEventHandler,
		producer:          producer,
		topics:            config.KafkaTopics,
	}
}

func (coordinator *Coordinator) HandleTransaction(event events.OrderPlaced, headers map[string]string) {
	orderItemsReserved, err := coordinator.orderEventHandler.Handle(event)
	if err != nil {
		// TOOD: create and send rollback message
	}
	coordinator.producer.Send(orderItemsReserved, headers, coordinator.topics.ItemsReservedTopic)
}
