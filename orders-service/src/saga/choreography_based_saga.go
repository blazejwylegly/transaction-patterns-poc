package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"github.com/google/uuid"
	"log"
	"time"
)

type Coordinator struct {
	producer         messaging.OrderProducer
	orderPlacedTopic string
}

func NewCoordinator(producer messaging.OrderProducer, kafkaConfig config.KafkaConfig) *Coordinator {
	return &Coordinator{producer, kafkaConfig.KafkaTopics.OrderRequestsTopic}
}

func (coordinator *Coordinator) BeginOrderPlacedTransaction(order models.Order) {
	log.Printf("Trying to initiate txn with orderId %s\n", order.OrderID.String())
	messageHeaders := map[string]string{
		messaging.StepNameHeader:             "ORDER_PLACED",
		messaging.StepExecutorHeader:         "ORDER_SERVICE",
		messaging.TransactionIdHeader:        uuid.New().String(),
		messaging.TransactionNameHeader:      "PRODUCT_PURCHASED",
		messaging.TransactionStartedAtHeader: time.Now().String(),
	}

	coordinator.producer.Send(order, messageHeaders, coordinator.orderPlacedTopic)
	log.Printf("Transaction with with orderId %s initiated successfully\n", order.OrderID.String())
}
