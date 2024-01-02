package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/events"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging/producer"
	"github.com/google/uuid"
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
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "ORDER_ITEMS_RESERVED",
		messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
		messaging.TransactionIdHeader:        headers[messaging.TransactionIdHeader],
		messaging.TransactionNameHeader:      headers[messaging.TransactionNameHeader],
		messaging.TransactionStartedAtHeader: headers[messaging.TransactionStartedAtHeader],
	}
	coordinator.producer.Send(orderItemsReserved, messageHeaders, coordinator.topics.ItemsReservedTopic)
}

func (coordinator *Coordinator) HandleRollback(event events.OrderFailed, headers map[string]string) {
	// TODO add compensating action
}
