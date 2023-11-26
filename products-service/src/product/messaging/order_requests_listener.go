package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/events"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/saga"
	"log"
	"sync"
)

type OrderListener struct {
	kafkaClient KafkaClient
	coordinator saga.Coordinator
	kafkaTopics config.KafkaTopics
}

func NewListener(client KafkaClient, config config.KafkaConfig, coordinator saga.Coordinator) *OrderListener {
	return &OrderListener{
		client,
		coordinator,
		config.KafkaTopics,
	}
}

// StartConsuming fetches all available partitions from kafka broker, and runs a separate
// goroutine to process the messages from each partition.
func (listener *OrderListener) StartConsuming() {
	partitions, _ := listener.kafkaClient.GetPartitions(listener.kafkaTopics.OrderRequestsTopic)
	var wg sync.WaitGroup
	for partition := range partitions {
		go listener.consumePartition(&wg, partition)
	}
}

// consumePartition is responsible for consuming specified partition by running two goroutines.
// One of them is used to continuously fetch messages from kafka and the second one is continuously handling them.
func (listener *OrderListener) consumePartition(wg *sync.WaitGroup, partition int32) {
	messagesChannel := make(chan *sarama.ConsumerMessage)
	defer close(messagesChannel)

	partitionConsumer, err := listener.kafkaClient.ConsumePartition(listener.kafkaTopics.OrderRequestsTopic, partition)
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

func (listener *OrderListener) processKafkaMessages(messagesChannel chan *sarama.ConsumerMessage) {
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

func parseEvent(msg *sarama.ConsumerMessage) (*events.OrderPlaced, error) {
	orderPlacedEvent := events.OrderPlaced{}
	err := json.Unmarshal(msg.Value, &orderPlacedEvent)
	if err != nil {
		fmt.Printf("Error parsing transaction message: %v\n", err)
		return nil, err
	}
	return &orderPlacedEvent, nil
}
