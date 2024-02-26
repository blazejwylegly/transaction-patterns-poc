package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging"
	"github.com/google/uuid"
	"log"
)

type Coordinator interface {
	HandleSagaEvent(application.OrderPlaced, map[string]string)
	HandleSagaRollbackEvent(application.OrderFailed, map[string]string)
}

type ChoreographyCoordinator struct {
	orderEventHandler application.OrderEventHandler
	producer          messaging.EventProducer
	topics            config.KafkaTopics
}

func NewChoreographyCoordinator(orderEventHandler application.OrderEventHandler,
	producer messaging.EventProducer,
	config config.KafkaConfig) *ChoreographyCoordinator {
	return &ChoreographyCoordinator{
		orderEventHandler: orderEventHandler,
		producer:          producer,
		topics:            config.KafkaTopics,
	}
}

func (coordinator *ChoreographyCoordinator) HandleSagaEvent(inputEvent application.OrderPlaced,
	inputHeaders map[string]string) {
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
		orderReservationFailed := &application.ItemReservationStatus{
			OrderID: inputEvent.OrderID,
			Details: err.Error(),
		}
		outputHeaders[messaging.StepResultHeader] = "FAILED"
		coordinator.producer.Send(orderReservationFailed, outputHeaders, coordinator.topics.TxnError)
		return
	}

	outputHeaders[messaging.StepResultHeader] = "SUCCESS"
	coordinator.producer.Send(orderItemsReserved, outputHeaders, coordinator.topics.ItemsReservedTopic)
}

func (coordinator *ChoreographyCoordinator) HandleSagaRollbackEvent(event application.OrderFailed,
	headers map[string]string) {
	if headers[messaging.StepExecutorHeader] != "PRODUCTS_SERVICE" {
		err := coordinator.orderEventHandler.HandleRollback(event)
		if err != nil {
			log.Printf("Error rollbacking order with id %s", event.OrderID)
			return
		}
		log.Printf("Successfully rollbacked order with id %s", event.OrderID)
	}
}
