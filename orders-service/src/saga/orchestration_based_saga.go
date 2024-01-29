package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"github.com/google/uuid"
	"log"
	"time"
)

type OrchestrationCoordinator struct {
	sagaLogger Logger
	producer   messaging.EventProducer
	topics     config.KafkaTopics
}

func NewOrchestrationCoordinator(sagaLogger Logger,
	producer messaging.EventProducer,
	kafkaConfig config.KafkaConfig) *OrchestrationCoordinator {
	return &OrchestrationCoordinator{sagaLogger, producer, kafkaConfig.KafkaTopics}
}

func (orchestrator *OrchestrationCoordinator) BeginOrderPlacedTransaction(order models.Order) {
	log.Printf("Trying to initiate txn with orderId %s\n", order.OrderID.String())
	_ = map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "ORDER_PLACED",
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepResultHeader:           "SUCCESS",
		messaging.TransactionIdHeader:        uuid.New().String(),
		messaging.TransactionNameHeader:      "PRODUCT_PURCHASED",
		messaging.TransactionStartedAtHeader: time.Now().String(),
	}
	orchestrator.RequestProductReservation(order)

	// request product reservation
	// verify result
	// request payment
	// verify result
	// produce order result
}

func (orchestrator *OrchestrationCoordinator) RequestProductReservation(order models.Order) {
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "ORDER_PLACED",
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepResultHeader:           "SUCCESS",
		messaging.TransactionIdHeader:        uuid.New().String(),
		messaging.TransactionNameHeader:      "PRODUCT_PURCHASED",
		messaging.TransactionStartedAtHeader: time.Now().String(),
	}
	orchestrator.producer.Send(order, messageHeaders, orchestrator.topics.OrderRequestsTopic)
	log.Printf("Transaction with with orderId %s initiated successfully\n", order.OrderID.String())
}

func (orchestrator *OrchestrationCoordinator) RollbackProductReservation(order models.Order) {

}

func (orchestrator *OrchestrationCoordinator) RequestPayment() {

}
