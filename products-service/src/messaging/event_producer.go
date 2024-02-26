package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"log"
	"time"
)

type EventProducer interface {
	Send(interface{}, map[string]string, string)
}

type SaramaEventProducer struct {
	kafkaProducer sarama.SyncProducer
}

func NewSaramaProducer(producer sarama.SyncProducer) *SaramaEventProducer {
	return &SaramaEventProducer{kafkaProducer: producer}
}

func (producer *SaramaEventProducer) Send(data interface{}, headers map[string]string, topic string) {
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

func createMessage[T any](message T, headers map[string]string, topicName string) (*sarama.ProducerMessage, error) {
	bytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println(err)
		return nil, nil
	}

	saramaHeaders := make([]sarama.RecordHeader, 0)
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
