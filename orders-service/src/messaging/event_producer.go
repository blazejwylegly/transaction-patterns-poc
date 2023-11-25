package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"log"
	"time"
)

type OrderProducer interface {
	Send(interface{}, map[string]string, string)
}

type SaramaOrderProducer struct {
	kafkaProducer     sarama.SyncProducer
	orderRequestTopic string
}

func NewSaramaProducer(producer sarama.SyncProducer, config config.KafkaConfig) *SaramaOrderProducer {
	return &SaramaOrderProducer{
		kafkaProducer:     producer,
		orderRequestTopic: config.KafkaTopics.OrderRequestsTopic,
	}
}

func (producer *SaramaOrderProducer) Send(data interface{}, headers map[string]string, topic string) {
	encodedMessage, err := createMessage(data, headers, topic)
	if err != nil {
		fmt.Println(err)
		return
	}

	partition, offset, err := producer.kafkaProducer.SendMessage(encodedMessage)
	if err != nil {
		log.Printf("Error producing message for txn id %s - %v}", headers[TransactionIdHeader], err)
		return
	}
	log.Printf(
		"Produced message {offset: '%d', partition: '%d', txnId: '%s'}",
		offset,
		partition,
		headers[TransactionIdHeader],
	)
}

func createMessage(data interface{}, headers map[string]string, topicName string) (*sarama.ProducerMessage, error) {
	bytes, err := json.Marshal(data)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	saramaHeaders := make([]sarama.RecordHeader, len(headers))
	for headerKey, headerValue := range headers {
		saramaHeaders = append(saramaHeaders, sarama.RecordHeader{Key: []byte(headerKey), Value: []byte(headerValue)})
	}

	producerMessage := &sarama.ProducerMessage{
		Topic:     topicName,
		Value:     sarama.ByteEncoder(bytes),
		Timestamp: time.Time{},
		Headers:   saramaHeaders,
	}
	return producerMessage, err
}
