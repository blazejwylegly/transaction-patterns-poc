package messaging

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"log"
	"os"
)

type KafkaClient struct {
	consumer     sarama.Consumer
	saramaConfig *sarama.Config
	kafkaConfig  config.KafkaConfig
}

func NewKafkaClient(kafkaConfig config.KafkaConfig) KafkaClient {
	saramaConfig := sarama.NewConfig()
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetOldest

	consumer, err := sarama.NewConsumer([]string{kafkaConfig.KafkaUrl}, nil)
	if err != nil {
		log.Printf("Error trying to connect to one of kafka brokers")
	}

	return KafkaClient{
		consumer:     consumer,
		saramaConfig: saramaConfig,
		kafkaConfig:  kafkaConfig,
	}
}

func (kafkaClient KafkaClient) GetPartitions(topic string) (chan int32, error) {

	partitions, err := kafkaClient.consumer.Partitions(topic)
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

func (kafkaClient KafkaClient) ConsumePartition(topic string, partition int32) (sarama.PartitionConsumer, error) {
	partitionConsumer, err := kafkaClient.consumer.ConsumePartition(topic, partition, sarama.OffsetOldest)
	if err != nil {
		fmt.Printf("Error creating partition consumer for partition %d: %v\n", partition, err)
		return nil, err
	}
	return partitionConsumer, nil
}
