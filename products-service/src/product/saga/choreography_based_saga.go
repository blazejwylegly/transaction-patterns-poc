package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/events"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging/producer"
	"github.com/google/uuid"
	"log"
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

func (coordinator *Coordinator) HandleTransaction(inputEvent events.OrderPlaced, inputHeaders map[string]string) {
	orderItemsReserved, err := coordinator.orderEventHandler.Handle(inputEvent)
	outputHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "ORDER_ITEMS_RESERVED",
		messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
		messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
		messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
		messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
	}

	if err != nil {
		log.Printf("Txn with id %s failed - initiating rollback", inputHeaders[messaging.TransactionIdHeader])
		paymentFailed := &events.OrderReservationFailed{
			OrderID: inputEvent.OrderID,
			Details: err.Error(),
		}
		outputHeaders[messaging.StepResultHeader] = "FAILED"
		coordinator.producer.Send(paymentFailed, outputHeaders, coordinator.topics.OrderFailedTopic)
		return
	}

	outputHeaders[messaging.StepResultHeader] = "SUCCESS"
	coordinator.producer.Send(orderItemsReserved, outputHeaders, coordinator.topics.ItemsReservedTopic)
}

func (coordinator *Coordinator) HandleRollback(event events.OrderFailed, headers map[string]string) {
	if headers[messaging.StepExecutorHeader] != "PRODUCTS_SERVICE" {
		err := coordinator.orderEventHandler.HandleRollback(event)
		if err != nil {
			log.Printf("Error rollbacking order with id %s", event.OrderID)
			return
		}
		log.Printf("Successfully rollbacked order with id %s", event.OrderID)
	}
}
