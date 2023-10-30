package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"log"
	"time"
)

type OrderRequestProducer struct {
	kafkaProducer     sarama.SyncProducer
	orderRequestTopic string
}

func NewProducer(producer sarama.SyncProducer, config config.Config) *OrderRequestProducer {
	return &OrderRequestProducer{
		kafkaProducer:     producer,
		orderRequestTopic: config.Kafka.Topics.OrderRequestsTopic,
	}
}

func (producer *OrderRequestProducer) Send(order models.Order) {
	message, err := jsonEncodedMessage(order, producer.orderRequestTopic)
	if err != nil {
		fmt.Println(err)
		return
	}

	partition, offset, err := producer.kafkaProducer.SendMessage(message)
	if err != nil {
		log.Printf("Error producing message for orderId %s - %v}", order.OrderID, err)
		return
	}
	log.Printf("Produced order message {offset: '%d', partition: '%d'}", offset, partition)
}

func jsonEncodedMessage(order models.Order, topicName string) (*sarama.ProducerMessage, error) {
	bytes, err := json.Marshal(order)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}
	message := &sarama.ProducerMessage{
		Topic:     topicName,
		Value:     sarama.ByteEncoder(bytes),
		Timestamp: time.Time{},
	}
	return message, err
}
