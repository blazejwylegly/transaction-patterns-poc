package listener

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/events"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/saga"
	"log"
	"sync"
)

type PaymentRequestListener struct {
	kafkaClient messaging.KafkaClient
	coordinator saga.Coordinator
	kafkaTopics config.KafkaTopics
}

func NewListener(client messaging.KafkaClient, config config.KafkaConfig, coordinator saga.Coordinator) *PaymentRequestListener {
	return &PaymentRequestListener{
		client,
		coordinator,
		config.KafkaTopics,
	}
}

// StartConsuming fetches all available partitions from kafka broker, and runs a separate
// goroutine to process the messages from each partition.
func (listener *PaymentRequestListener) StartConsuming() {
	partitions, _ := listener.kafkaClient.GetPartitions(listener.kafkaTopics.ItemsReservedTopic)
	var wg sync.WaitGroup
	for partition := range partitions {
		go listener.consumePartition(&wg, partition)
	}
}

// consumePartition is responsible for consuming specified partition by running two goroutines.
// One of them is used to continuously fetch messages from kafka and the second one is continuously handling them.
func (listener *PaymentRequestListener) consumePartition(wg *sync.WaitGroup, partition int32) {
	messagesChannel := make(chan *sarama.ConsumerMessage)
	defer close(messagesChannel)

	partitionConsumer, err := listener.kafkaClient.ConsumePartition(listener.kafkaTopics.ItemsReservedTopic, partition)
	if err != nil {
		log.Printf("Error creating consumer for partition %d: %v", partition, err)
	}

	wg.Add(1)
	go func(pc sarama.PartitionConsumer) {
		defer wg.Done()
		for message := range pc.Messages() {
			messagesChannel <- message
		}
	}(partitionConsumer)

	defer func(partitionConsumer sarama.PartitionConsumer) {
		err := partitionConsumer.Close()
		if err != nil {
			fmt.Printf("Error closing partition consumer for partition %s %v\n", partition, err)
		}
	}(partitionConsumer)

	go listener.processKafkaMessages(messagesChannel)
	select {}
}

func (listener *PaymentRequestListener) processKafkaMessages(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		txMessage, err := parseEvent(msg)
		headers := parseHeaders(msg.Headers)
		listener.coordinator.HandleTransaction(*txMessage, headers)
		if err != nil {
			fmt.Printf("Error parsing msg { partition:'%d', offset:'%d' }: %v\n", msg.Partition, msg.Offset, err)
		}
	}
}

func parseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}

func parseEvent(msg *sarama.ConsumerMessage) (*events.PaymentRequested, error) {
	paymentRequested := events.PaymentRequested{}
	err := json.Unmarshal(msg.Value, &paymentRequested)
	if err != nil {
		fmt.Printf("Error parsing transaction message: %v\n", err)
		return nil, err
	}
	return &paymentRequested, nil
}
