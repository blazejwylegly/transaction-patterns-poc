package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/application"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/google/uuid"
	"log"
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
		messaging.TransactionStatusHeader:    "PENDING",
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "PRODUCT_RESERVATION_REQUESTED",
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           "PENDING",
	}
	orchestrator.producer.Send(order, messageHeaders, orchestrator.topics.ItemReservationRequest)
	log.Printf("Requested reservation for order %s\n", order.OrderID.String())
}

func (orchestrator *OrchestrationCoordinator) HandleItemReservationStatusEvent(
	itemReservationStatus application.ItemReservationStatus,
	context Context) {
	log.Printf("Handling ItemsReserved for order %s\n", itemReservationStatus.OrderID.String())

	if itemReservationStatus.Status == "success" {
		paymentRequest := application.PaymentRequest{
			OrderID:    itemReservationStatus.OrderID,
			CustomerID: itemReservationStatus.CustomerID,
			TotalCost:  itemReservationStatus.TotalCost,
		}
		orchestrator.requestPayment(paymentRequest, context)
		log.Printf("Item reservation for order %s successful. Requesting order payment\n",
			itemReservationStatus.OrderID.String())
	} else {
		log.Printf("Item reservation for order %s failed. Aborting order processing\n",
			itemReservationStatus.OrderID.String())
		orderFailed := application.OrderFailed{
			OrderID:    itemReservationStatus.OrderID,
			CustomerID: itemReservationStatus.CustomerID,
			Details:    "Failed to reserve requested items",
		}
		orchestrator.terminateOrder(context, orderFailed)
	}

}

func (orchestrator *OrchestrationCoordinator) requestPayment(request application.PaymentRequest, context Context) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               context.PreviousStepId.String(),
		messaging.StepNameHeader:             "ORDER_PAYMENT_REQUESTED",
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepStatusHeader:           "REQUESTED",
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    "PENDING",
	}
	orchestrator.producer.Send(request, messageHeaders, orchestrator.topics.PaymentRequest)
}

func (orchestrator *OrchestrationCoordinator) HandlePaymentStatusEvent(paymentStatus application.PaymentStatus,
	context Context) {

	log.Printf("Handling ItemsReserved for order %s\n", paymentStatus.OrderID.String())

	if paymentStatus.Status == "success" {
		orderCompleted := application.OrderCompleted{
			OrderID:    paymentStatus.OrderID,
			CustomerID: paymentStatus.CustomerID,
			Details:    "",
		}
		orchestrator.completeOrder(context, orderCompleted)
		log.Printf("Order %s completed", orderCompleted.OrderID.String())
	} else {
		log.Printf("Payment for order %s failed. Requesting global rollback of all steps.",
			paymentStatus.OrderID.String())
		paymentFailed := application.PaymentFailed{
			OrderID: paymentStatus.OrderID,
		}
		orchestrator.terminateOrder(context, paymentFailed)
	}

}

func (orchestrator *OrchestrationCoordinator) requestProductReservationRollback(orderId uuid.UUID, context Context) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               context.PreviousStepId.String(),
		messaging.StepNameHeader:             "PRODUCT_RESERVATION_ROLLBACK_REQUESTED",
		messaging.StepExecutorHeader:         "PAYMENT_SERVICE",
		messaging.StepStatusHeader:           "REQUESTED",
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    "PENDING",
	}
	paymentFailedEvent := application.PaymentFailed{
		OrderID: orderId,
	}
	orchestrator.producer.Send(paymentFailedEvent, messageHeaders, orchestrator.topics.ItemReservationStatus)
}

func (orchestrator *OrchestrationCoordinator) terminateOrder(context Context, failedEvent interface{}) {
	messageHeaders := map[string]string{
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    "FAILED",
	}
	orchestrator.producer.Send(failedEvent, messageHeaders, orchestrator.topics.OrderStatus)
}

func (orchestrator *OrchestrationCoordinator) completeOrder(context Context, orderCompleted application.OrderCompleted) {
	messageHeaders := map[string]string{
		messaging.TransactionIdHeader:        context.TransactionId.String(),
		messaging.TransactionNameHeader:      context.TransactionName,
		messaging.TransactionStartedAtHeader: context.TransactionStartedAt,
		messaging.TransactionStatusHeader:    "COMPLETED",
	}
	orchestrator.producer.Send(orderCompleted, messageHeaders, orchestrator.topics.OrderStatus)
}
