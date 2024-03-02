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

func (coordinator *OrchestrationCoordinator) HandleSagaEvent(inputEvent application.UpdateRequested,
	inputHeaders map[string]string) {

	if inputEvent.IsRollback {
		coordinator.HandleSagaRollbackEvent(application.OrderFailed{
			OrderID: inputEvent.OrderID,
		}, inputHeaders)
	} else {
		coordinator.handleItemUpdateEvent(inputEvent, inputHeaders)
	}

}
func (coordinator *OrchestrationCoordinator) handleItemUpdateEvent(inputEvent application.UpdateRequested,
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

func (coordinator *OrchestrationCoordinator) HandleSagaRollbackEvent(inputEvent application.OrderFailed,
	inputHeaders map[string]string) {
	rollbackEvent := application.OrderFailed{
		OrderID: inputEvent.OrderID,
	}
	err := coordinator.orderEventHandler.HandleRollback(rollbackEvent)
	if err != nil {
		log.Printf("Items reservation rollback for order %s failed", inputEvent.OrderID)
		outputHeaders := map[string]string{
			messaging.StepIdHeader:               uuid.New().String(),
			messaging.StepNameHeader:             "INVENTORY_UPDATE_ROLLBACK_REQUESTED",
			messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
			messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
			messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
			messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
		}
		rollbackFailedEvent := &application.ItemReservationStatus{
			OrderID: inputEvent.OrderID,
			Details: err.Error(),
			Status:  "ROLLBACK_FAILED",
		}
		outputHeaders[messaging.StepResultHeader] = "FAILED"
		coordinator.producer.Send(rollbackFailedEvent, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	} else {
		log.Printf("Items reservation rollback for order %s syccess", inputEvent.OrderID)
		outputHeaders := map[string]string{
			messaging.StepIdHeader:               uuid.New().String(),
			messaging.StepNameHeader:             "INVENTORY_UPDATE_ROLLBACK_REQUESTED",
			messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
			messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
			messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
			messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
		}
		rollbackSuccessEvent := &application.ItemReservationStatus{
			OrderID: inputEvent.OrderID,
			Status:  "ROLLBACK_SUCCESS",
		}
		outputHeaders[messaging.StepResultHeader] = "SUCCESS"
		coordinator.producer.Send(rollbackSuccessEvent, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	}
}
