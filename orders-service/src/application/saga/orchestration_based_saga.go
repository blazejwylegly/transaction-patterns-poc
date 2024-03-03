package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/application"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/google/uuid"
	"log"
	"strings"
	"time"
)

type OrchestrationCoordinator struct {
	sagaLogger Logger
	producer   messaging.EventProducer
	topics     config.OrchestrationTopics
}

func NewOrchestrationCoordinator(sagaLogger Logger,
	producer messaging.EventProducer,
	kafkaConfig config.KafkaConfig) *OrchestrationCoordinator {
	return &OrchestrationCoordinator{sagaLogger, producer, kafkaConfig.OrchestrationTopics}
}

func (orchestrator *OrchestrationCoordinator) BeginOrderPlacedTransaction(order application.Order) {
	log.Printf("Trying to initiate txn with orderId %s\n", order.OrderID.String())
	context := Context{
		TransactionId:        uuid.New(),
		TransactionName:      "PRODUCT_PURCHASED",
		TransactionStartedAt: time.Now().String(),
	}
	orchestrator.requestProductReservation(order, context)
}

func (orchestrator *OrchestrationCoordinator) requestProductReservation(order application.Order, context Context) {
	messageHeaders := map[string]string{
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    messaging.TxnStatusPending,
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             messaging.ItemReservationRequestedStep,
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           messaging.StepStatusPending,
		messaging.StepStartedAtHeader:        time.Now().String(),
	}
	orchestrator.producer.Send(order, messageHeaders, orchestrator.topics.InventoryUpdateRequest)
	log.Printf("Requested reservation for order %s\n", order.OrderID.String())
}

func (orchestrator *OrchestrationCoordinator) HandleItemReservationStatusEvent(
	itemReservationStatus application.ItemReservationStatus,
	context Context) {
	log.Printf("Handling ItemsReserved for order %s\n", itemReservationStatus.OrderID.String())

	if strings.EqualFold(context.PreviousStepStatus, messaging.StepStatusSuccess) {
		paymentRequest := application.PaymentRequest{
			OrderID:    itemReservationStatus.OrderID,
			CustomerID: itemReservationStatus.CustomerID,
			TotalCost:  itemReservationStatus.TotalCost,
		}
		orchestrator.requestPayment(paymentRequest, context)
		log.Printf("Item reservation for order %s successful. Requesting order payment\n",
			itemReservationStatus.OrderID.String())
		return
	}

	if strings.EqualFold(context.PreviousStepStatus, messaging.StepStatusFailed) {
		log.Printf("Item reservation for order %s failed. Aborting order processing\n",
			itemReservationStatus.OrderID.String())
		orderFailed := application.OrderFailed{
			OrderID:    itemReservationStatus.OrderID,
			CustomerID: itemReservationStatus.CustomerID,
			Details:    "Failed to reserve requested items",
		}
		orchestrator.terminateOrder(context, orderFailed)
		return
	}
}

func (orchestrator *OrchestrationCoordinator) HandleItemReservationRollbackStatusEvent(
	itemReservationRollbackStatus application.ItemReservationStatus,
	context Context) {
	log.Printf("Handling ItemsReserved for order %s\n", itemReservationRollbackStatus.OrderID.String())

	if strings.EqualFold(context.PreviousStepStatus, messaging.StepStatusSuccess) {
		log.Printf("Item reservation rollback for order %s successful. Terminating order.\n",
			itemReservationRollbackStatus.OrderID.String())
		orderFailed := application.OrderFailed{
			OrderID:    itemReservationRollbackStatus.OrderID,
			CustomerID: itemReservationRollbackStatus.CustomerID,
			Details:    "Failed to process payment for order. Reserved items have been released",
		}
		orchestrator.terminateOrder(context, orderFailed)
		return
	}

	if strings.EqualFold(context.PreviousStepStatus, messaging.StepStatusFailed) {
		log.Printf("Item reservation for order %s failed. Aborting order processing\n",
			itemReservationRollbackStatus.OrderID.String())
		orderFailed := application.OrderFailed{
			OrderID:    itemReservationRollbackStatus.OrderID,
			CustomerID: itemReservationRollbackStatus.CustomerID,
			Details:    "Failed to release requested items. Manual intervention required!",
		}
		orchestrator.terminateOrder(context, orderFailed)
		return
	}

}

func (orchestrator *OrchestrationCoordinator) requestPayment(request application.PaymentRequest, context Context) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             messaging.PaymentRequestedStep,
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           messaging.StepStatusPending,
		messaging.StepStartedAtHeader:        time.Now().String(),
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    messaging.TxnStatusPending,
	}
	orchestrator.producer.Send(request, messageHeaders, orchestrator.topics.PaymentRequest)
}

func (orchestrator *OrchestrationCoordinator) HandlePaymentStatusEvent(paymentStatus application.PaymentStatus,
	context Context) {

	log.Printf("Handling ItemsReserved for order %s\n", paymentStatus.OrderID.String())

	if strings.EqualFold(paymentStatus.Status, messaging.StepStatusSuccess) {
		orderCompleted := application.OrderCompleted{
			OrderID:    paymentStatus.OrderID,
			CustomerID: paymentStatus.CustomerID,
			Details:    "Order completed successfully",
		}
		orchestrator.completeOrder(context, orderCompleted)
		log.Printf("Order %s completed", orderCompleted.OrderID.String())
	} else {
		log.Printf("Payment for order %s failed. Requesting global rollback of all steps.",
			paymentStatus.OrderID.String())
		orchestrator.requestProductReservationRollback(paymentStatus.OrderID, context)
	}

}

func (orchestrator *OrchestrationCoordinator) requestProductReservationRollback(orderId uuid.UUID, context Context) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             messaging.ItemReservationRollbackRequestedStep,
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           messaging.StepStatusPending,
		messaging.StepStartedAtHeader:        time.Now().String(),
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    messaging.TxnStatusRollbackInProgress,
	}
	paymentFailedEvent := application.ItemReservationRollbackRequest{
		OrderID: orderId,
	}
	orchestrator.producer.Send(paymentFailedEvent, messageHeaders, orchestrator.topics.InventoryUpdateRequest)
}

func (orchestrator *OrchestrationCoordinator) terminateOrder(context Context, failedEvent interface{}) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             messaging.OrderCancelledStep,
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           messaging.StepStatusSuccess,
		messaging.StepStartedAtHeader:        time.Now().String(),
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    messaging.TxnStatusFailed,
	}
	orchestrator.producer.Send(failedEvent, messageHeaders, orchestrator.topics.OrderStatus)
}

func (orchestrator *OrchestrationCoordinator) completeOrder(context Context, orderCompleted application.OrderCompleted) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             messaging.OrderCompletedStep,
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           messaging.StepStatusSuccess,
		messaging.StepStartedAtHeader:        time.Now().String(),
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    messaging.TxnStatusSuccess,
	}
	orchestrator.producer.Send(orderCompleted, messageHeaders, orchestrator.topics.OrderStatus)
}
