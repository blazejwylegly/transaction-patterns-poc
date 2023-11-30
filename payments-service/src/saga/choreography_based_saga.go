package saga

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/application"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging/producer"
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
	coordinator.producer.Send(paymentProcessed, headers, coordinator.topics.OrderResultsTopic)
}
