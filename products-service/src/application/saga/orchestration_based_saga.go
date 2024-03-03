package saga

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging"
	"github.com/google/uuid"
	"log"
	"time"
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

func (coordinator *OrchestrationCoordinator) HandleSagaEvent(inputEvent application.InventoryUpdateRequest,
	inputHeaders map[string]string) {

	if inputHeaders[messaging.StepNameHeader] == messaging.ItemReservationRollbackRequestedStep {
		coordinator.HandleSagaRollbackEvent(application.OrderFailed{
			OrderID: inputEvent.OrderID,
		}, inputHeaders)
	} else if inputHeaders[messaging.StepNameHeader] == messaging.ItemReservationRequestedStep {
		coordinator.handleInventoryUpdateEvent(inputEvent, inputHeaders)
	} else {
		log.Fatalf("Orchestration - uncrecognized step!")
	}

}
func (coordinator *OrchestrationCoordinator) handleInventoryUpdateEvent(inputEvent application.InventoryUpdateRequest,
	inputHeaders map[string]string) {
	itemsReserved, err := coordinator.orderEventHandler.Handle(inputEvent)
	outputHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             inputHeaders[messaging.StepNameHeader],
		messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
		messaging.StepStartedAtHeader:        time.Now().String(),
		messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
		messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
		messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
	}
	if err != nil {
		log.Printf("Items reservation for order %s failed", inputEvent.OrderID)
		itemsReservationFailed := &application.ItemReservationStatus{
			OrderID:    inputEvent.OrderID,
			CustomerID: itemsReserved.CustomerID,
			Details:    err.Error(),
			Status:     messaging.StepStatusFailed,
		}
		outputHeaders[messaging.StepStatusHeader] = messaging.StepStatusFailed
		coordinator.producer.Send(itemsReservationFailed, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	} else {
		itemReservationSuccessful := &application.ItemReservationStatus{
			OrderID:    itemsReserved.OrderID,
			CustomerID: itemsReserved.CustomerID,
			TotalCost:  itemsReserved.TotalCost,
			Status:     messaging.StepStatusSuccess,
		}
		log.Printf("Items reservation for order %s successful", inputEvent.OrderID)
		outputHeaders[messaging.StepStatusHeader] = messaging.StepStatusSuccess
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
			messaging.StepNameHeader:             inputHeaders[messaging.StepNameHeader],
			messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
			messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
			messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
			messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
			messaging.StepStartedAtHeader:        time.Now().String(),
		}
		rollbackFailedEvent := &application.ItemReservationStatus{
			OrderID: inputEvent.OrderID,
			Details: err.Error(),
			Status:  messaging.StepStatusFailed,
		}
		outputHeaders[messaging.StepStatusHeader] = messaging.StepStatusFailed
		coordinator.producer.Send(rollbackFailedEvent, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	} else {
		log.Printf("Items reservation rollback for order %s success", inputEvent.OrderID)
		outputHeaders := map[string]string{
			messaging.StepIdHeader:               uuid.New().String(),
			messaging.StepNameHeader:             inputHeaders[messaging.StepNameHeader],
			messaging.StepExecutorHeader:         "PRODUCTS_SERVICE",
			messaging.TransactionIdHeader:        inputHeaders[messaging.TransactionIdHeader],
			messaging.TransactionNameHeader:      inputHeaders[messaging.TransactionNameHeader],
			messaging.TransactionStartedAtHeader: inputHeaders[messaging.TransactionStartedAtHeader],
			messaging.StepStartedAtHeader:        time.Now().String(),
		}
		rollbackSuccessEvent := &application.ItemsReleased{
			OrderID: inputEvent.OrderID,
		}
		outputHeaders[messaging.StepStatusHeader] = messaging.StepStatusSuccess
		coordinator.producer.Send(rollbackSuccessEvent, outputHeaders, coordinator.topics.InventoryUpdateStatus)
	}
}
