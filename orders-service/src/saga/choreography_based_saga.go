package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"github.com/google/uuid"
	"log"
	"time"
)

type Coordinator interface {
	BeginOrderPlacedTransaction(order models.Order)
}

type ChoreographyCoordinator struct {
	producer         messaging.EventProducer
	orderPlacedTopic string
}

func NewChoreographyCoordinator(producer messaging.EventProducer,
	kafkaConfig config.KafkaConfig) *ChoreographyCoordinator {
	return &ChoreographyCoordinator{producer, kafkaConfig.KafkaTopics.OrderRequestsTopic}
}

func (overseer *ChoreographyCoordinator) BeginOrderPlacedTransaction(order models.Order) {
	log.Printf("Trying to initiate txn with orderId %s\n", order.OrderID.String())
	messageHeaders := map[string]string{
		messaging.StepIdHeader:               uuid.New().String(),
		messaging.StepNameHeader:             "ORDER_PLACED",
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.StepResultHeader:           "SUCCESS",
		messaging.TransactionIdHeader:        uuid.New().String(),
		messaging.TransactionNameHeader:      "PRODUCT_PURCHASED",
		messaging.TransactionStartedAtHeader: time.Now().String(),
	}

	overseer.producer.Send(order, messageHeaders, overseer.orderPlacedTopic)
	log.Printf("Transaction with with orderId %s initiated successfully\n", order.OrderID.String())
}
