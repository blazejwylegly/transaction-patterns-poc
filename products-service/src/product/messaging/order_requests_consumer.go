package messaging

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
)

type OrderListener struct {
	kafkaClient        KafkaClient
	orderRequestsTopic string
}

func NewListener(client KafkaClient, config config.KafkaConfig) OrderListener {
	return OrderListener{
		client,
		config.KafkaTopics.OrderRequestsTopic,
	}
}

func (listener OrderListener) StartConsuming() {
	partitions, _ := listener.kafkaClient.GetPartitions(listener.orderRequestsTopic)
	for partition := range partitions {
		go listener.consumePartition(partition)
	}
	select {}
}

func (listener OrderListener) consumePartition(partition int32) {
	messagesChannel := make(chan *sarama.ConsumerMessage)
	defer close(messagesChannel)
	go listener.kafkaClient.ConsumePartition(listener.orderRequestsTopic, partition, messagesChannel)
	for message := range messagesChannel {
		fmt.Printf("Consumer for partiton %d received message %s \n", partition, message.Value)
	}
}
