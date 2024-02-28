package saga

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/application"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging/producer"
	"github.com/google/uuid"
	"log"
)

type OrchestrationCoordinator struct {
	eventHandler application.PaymentRequestedHandler
	producer     producer.EventProducer
	topics       config.KafkaTopics
}

func NewOrchestrationCoordinator(paymentEventHandler application.PaymentRequestedHandler,
	producer producer.EventProducer,
	config config.KafkaConfig) *OrchestrationCoordinator {
	return &OrchestrationCoordinator{
		eventHandler: paymentEventHandler,
		producer:     producer,
		topics:       config.KafkaTopics,
	}
}

func (coordinator *OrchestrationCoordinator) HandleSagaEvent(inputEvent events.PaymentRequested,
	headers map[string]string) {
	paymentProcessed, err := coordinator.eventHandler.Handle(inputEvent)
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "PAYMENT_COMPLETED",
		messaging.StepExecutorHeader:         "PAYMENTS_SERVICE",
		messaging.TransactionIdHeader:        headers[messaging.TransactionIdHeader],
		messaging.TransactionNameHeader:      headers[messaging.TransactionNameHeader],
		messaging.TransactionStartedAtHeader: headers[messaging.TransactionStartedAtHeader],
	}
	if err != nil {
		log.Printf("Txn with id %s failed - initiating rollback", headers[messaging.TransactionIdHeader])
		paymentFailed := &events.PaymentFailed{
			OrderID:    inputEvent.OrderID,
			CustomerID: inputEvent.CustomerID,
			Details:    err.Error(),
		}
		messageHeaders[messaging.StepResultHeader] = "FAILED"
		coordinator.producer.Send(paymentFailed, messageHeaders, coordinator.topics.PaymentStatus)
	} else {
		messageHeaders[messaging.StepResultHeader] = "SUCCESS"
		coordinator.producer.Send(paymentProcessed, messageHeaders, coordinator.topics.PaymentStatus)
	}

}
