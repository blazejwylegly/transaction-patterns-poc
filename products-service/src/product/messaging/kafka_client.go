package messaging

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"log"
	"os"
)

type KafkaClient struct {
	saramaConfig *sarama.Config
	kafkaConfig  config.KafkaConfig
}

func NewKafkaClient(kafkaConfig config.KafkaConfig) KafkaClient {
	saramaConfig := sarama.NewConfig()
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest
	return KafkaClient{
		saramaConfig: saramaConfig,
		kafkaConfig:  kafkaConfig,
	}
}

func (kafkaClient KafkaClient) GetPartitions(topic string) (chan int32, error) {
	connection, err := sarama.NewClient([]string{kafkaClient.kafkaConfig.KafkaUrl}, nil)
	if err != nil {
		log.Printf("Error trying to connect to one of kafka brokers")
		return nil, err
	}

	partitions, err := connection.Partitions(topic)
	if err != nil {
		log.Printf("Error trying to obtain partitions for topic %s", topic)
		return nil, err
	}

	partitionsChannel := make(chan int32)

	go func() {
		defer close(partitionsChannel)
		for _, partition := range partitions {
			partitionsChannel <- partition
		}
	}()

	return partitionsChannel, nil
}

func (kafkaClient KafkaClient) ConsumePartition(topic string,
	partition int32,
	messageChannel chan<- *sarama.ConsumerMessage) {
	consumer, err := sarama.NewConsumer([]string{kafkaClient.kafkaConfig.KafkaUrl}, kafkaClient.saramaConfig)
	if err != nil {
		fmt.Printf("Error creating base consumer for partition %d: %v\n", partition, err)
		return
	}
	defer func(consumer sarama.Consumer) {
		err := consumer.Close()
		if err != nil {
			fmt.Printf("Error closing base consumer for partition %d: %v\n", partition, err)
		}
	}(consumer)

	partitionConsumer, err := consumer.ConsumePartition(topic, partition, sarama.OffsetOldest)
	if err != nil {
		fmt.Printf("Error creating partition consumer for partition %d: %v\n", partition, err)
		return
	}
	defer func(partitionConsumer sarama.PartitionConsumer) {
		err := partitionConsumer.Close()
		if err != nil {
			fmt.Printf("Error closing partition consumer for partition %d: %v\n", partition, err)
		}
	}(partitionConsumer)

	for {
		select {
		case message := <-partitionConsumer.Messages():
			messageChannel <- message
		}
	}
}
