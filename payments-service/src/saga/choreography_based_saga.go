package saga

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/application"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging/producer"
	"github.com/google/uuid"
)

type Coordinator struct {
	paymentRequestedHandler application.PaymentRequestedHandler
	producer                producer.EventProducer
	topics                  config.KafkaTopics
}

func NewCoordinator(itemsReservedHandler application.PaymentRequestedHandler,
	producer producer.EventProducer,
	config config.KafkaConfig) *Coordinator {
	return &Coordinator{
		paymentRequestedHandler: itemsReservedHandler,
		producer:                producer,
		topics:                  config.KafkaTopics,
	}
}

func (coordinator *Coordinator) HandleTransaction(event events.PaymentRequested, headers map[string]string) {
	paymentProcessed, err := coordinator.paymentRequestedHandler.Handle(event)
	if err != nil {
		// TOOD: create and send rollback message
	}
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "PAYMENT_PROCESSED",
		messaging.StepExecutorHeader:         "PAYMENTS_SERVICE",
		messaging.TransactionIdHeader:        headers[messaging.TransactionIdHeader],
		messaging.TransactionNameHeader:      headers[messaging.TransactionNameHeader],
		messaging.TransactionStartedAtHeader: headers[messaging.TransactionStartedAtHeader],
	}
	coordinator.producer.Send(paymentProcessed, messageHeaders, coordinator.topics.OrderResultsTopic)
}
