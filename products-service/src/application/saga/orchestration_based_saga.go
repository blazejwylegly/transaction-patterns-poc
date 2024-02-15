package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging/producer"
	"github.com/google/uuid"
	"log"
)

type OrchestrationCoordinator struct {
	orderEventHandler application.OrderEventHandler
	producer          producer.EventProducer
	topics            config.KafkaTopics
}

func NewOrchestrationCoordinator(orderEventHandler application.OrderEventHandler,
	producer producer.EventProducer,
	config config.KafkaConfig) *OrchestrationCoordinator {
	return &OrchestrationCoordinator{
		orderEventHandler: orderEventHandler,
		producer:          producer,
		topics:            config.KafkaTopics,
	}
}

func (coordinator *OrchestrationCoordinator) HandleTransaction(inputEvent application.OrderPlaced, inputHeaders map[string]string) {
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
		orderReservationFailed := &application.OrderReservationFailed{
			OrderID: inputEvent.OrderID,
			Details: err.Error(),
		}
		outputHeaders[messaging.StepResultHeader] = "FAILED"
		coordinator.producer.Send(orderReservationFailed, outputHeaders, coordinator.topics.OrderFailedTopic)
		return
	}

	outputHeaders[messaging.StepResultHeader] = "SUCCESS"
	coordinator.producer.Send(orderItemsReserved, outputHeaders, coordinator.topics.ItemsReservedTopic)
}

func (coordinator *Coordinator) HandleRollback(event application.OrderFailed, headers map[string]string) {
	if headers[messaging.StepExecutorHeader] != "PRODUCTS_SERVICE" {
		err := coordinator.orderEventHandler.HandleRollback(event)
		if err != nil {
			log.Printf("Error rollbacking order with id %s", event.OrderID)
			return
		}
		log.Printf("Successfully rollbacked order with id %s", event.OrderID)
	}
}
