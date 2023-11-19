package messaging

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	orderModels "github.com/blazejwylegly/transactions-poc/products-service/src/order"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/service"
	"log"
	"sync"
)

type OrderListener struct {
	kafkaClient          KafkaClient
	orderRequestsTopic   string
	orderRequestsHandler service.OrderRequestHandler
}

func NewListener(client KafkaClient, config config.KafkaConfig, handler service.OrderRequestHandler) OrderListener {
	return OrderListener{
		client,
		config.KafkaTopics.OrderRequestsTopic,
		handler,
	}
}

// StartConsuming fetches all available partitions from kafka broker, and runs a separate
// goroutine to process the messages from each partition.
func (listener OrderListener) StartConsuming() {
	partitions, _ := listener.kafkaClient.GetPartitions(listener.orderRequestsTopic)
	var wg sync.WaitGroup
	for partition := range partitions {
		go listener.consumePartition(&wg, partition)
	}
}

// consumePartition is responsible for consuming specified partition by running two goroutines.
// One of them is used to continuously fetch messages from kafka and the second one is continuously handling them.
func (listener OrderListener) consumePartition(wg *sync.WaitGroup, partition int32) {
	messagesChannel := make(chan *sarama.ConsumerMessage)
	defer close(messagesChannel)

	partitionConsumer, err := listener.kafkaClient.ConsumePartition(listener.orderRequestsTopic, partition)
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

	go listener.processKafkaMessages(partition, messagesChannel)
	select {}
}

func (listener OrderListener) processKafkaMessages(partition int32, messagesChannel chan *sarama.ConsumerMessage) {
	for message := range messagesChannel {
		orderRequest := &orderModels.OrderRequestDto{}
		err := json.Unmarshal(message.Value, orderRequest)
		if err != nil {
			fmt.Printf("Error handling message { partition:'%d', offset:'%d' }: %v\n", partition, message.Offset, err)
		}
		listener.orderRequestsHandler.Handle(orderRequest)
	}
}
