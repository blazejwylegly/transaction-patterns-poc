package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging"
	"github.com/google/uuid"
	"log"
)

type OrchestrationCoordinator struct {
	orderEventHandler application.OrderEventHandler
	producer          messaging.EventProducer
	topics            config.KafkaTopics
}

func NewOrchestrationCoordinator(orderEventHandler application.OrderEventHandler,
	producer messaging.EventProducer,
	config config.KafkaConfig) *OrchestrationCoordinator {
	return &OrchestrationCoordinator{
		orderEventHandler: orderEventHandler,
		producer:          producer,
		topics:            config.KafkaTopics,
	}
}

func (coordinator *OrchestrationCoordinator) HandleSagaEvent(inputEvent application.OrderPlaced,
	inputHeaders map[string]string) {
	itemsReserved, err := coordinator.orderEventHandler.Handle(inputEvent)
	outputHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "INVENTORY_UPDATE_REQUESTED",
		messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
		messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
		messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
		messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
	}

	if err != nil {
		log.Printf("Items reservation for order %s failed", inputEvent.OrderID)
		itemsReservationFailed := &application.ItemReservationStatus{
			OrderID: inputEvent.OrderID,
			Details: err.Error(),
			Status:  "FAILED",
		}
		outputHeaders[messaging.StepResultHeader] = "FAILED"
		coordinator.producer.Send(itemsReservationFailed, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	} else {
		itemReservationSuccessful := &application.ItemReservationStatus{
			OrderID:    itemsReserved.OrderID,
			CustomerID: itemsReserved.CustomerID,
			TotalCost:  itemsReserved.TotalCost,
			Status:     "SUCCESS",
		}
		log.Printf("Items reservation for order %s successful", inputEvent.OrderID)
		outputHeaders[messaging.StepResultHeader] = "SUCCESS"
		coordinator.producer.Send(itemReservationSuccessful, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	}
}

func (coordinator *OrchestrationCoordinator) HandleSagaRollbackEvent(event application.OrderFailed,
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
