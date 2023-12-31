package listener

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging"
	"log"
	"sync"
)

type MessageProcessor interface {
	process(messagesChannel chan *sarama.ConsumerMessage)
}

type TopicListener struct {
	kafkaClient messaging.KafkaClient
	processor   MessageProcessor
	topic       string
}

func NewListener(client messaging.KafkaClient, processor MessageProcessor, topic string) *TopicListener {
	return &TopicListener{
		client,
		processor,
		topic,
	}
}

// StartConsuming fetches all available partitions from kafka broker, and runs a separate
// goroutine to process the messages from each partition.
func (listener *TopicListener) StartConsuming() {
	partitions, _ := listener.kafkaClient.GetPartitions(listener.topic)
	var wg sync.WaitGroup
	for partition := range partitions {
		go listener.consumePartition(&wg, partition)
	}
}

// consumePartition is responsible for consuming specified partition by running two goroutines.
// One of them is used to continuously fetch messages from kafka and the second one is continuously handling them.
func (listener *TopicListener) consumePartition(wg *sync.WaitGroup, partition int32) {
	messagesChannel := make(chan *sarama.ConsumerMessage)
	defer close(messagesChannel)

	partitionConsumer, err := listener.kafkaClient.ConsumePartition(listener.topic, partition)
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

	go listener.processor.process(messagesChannel)
	select {}
}
