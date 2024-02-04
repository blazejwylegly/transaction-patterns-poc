package messaging

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"log"
	"os"
	"time"
)

type KafkaClient struct {
	consumer     sarama.Consumer
	producer     sarama.SyncProducer
	saramaConfig *sarama.Config
	kafkaConfig  config.KafkaConfig
}

func NewKafkaClient(kafkaConfig config.KafkaConfig) *KafkaClient {
	saramaConfig := sarama.NewConfig()

	// Commons
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	saramaConfig.Metadata.Retry.Max = 10
	saramaConfig.Metadata.Retry.Backoff = time.Second * 5

	// Consumer config
	saramaConfig.Consumer.Offsets.Initial = sarama.OffsetNewest

	// Producer config
	saramaConfig.Producer.Flush.Frequency = time.Duration(kafkaConfig.KafkaFlushFrequencyMs) * time.Millisecond
	saramaConfig.Producer.Compression = sarama.CompressionGZIP
	saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = true

	return &KafkaClient{
		saramaConfig: saramaConfig,
		kafkaConfig:  kafkaConfig,
	}
}

func (kafkaClient *KafkaClient) NewProducer() (*sarama.SyncProducer, error) {
	producer, err := sarama.NewSyncProducer([]string{kafkaClient.kafkaConfig.KafkaUrl}, kafkaClient.saramaConfig)
	if err != nil {
		log.Printf("Error creating producer: %v", err)
		return nil, err
	}
	kafkaClient.producer = producer
	return &producer, nil
}

func (kafkaClient *KafkaClient) NewConsumer() (*sarama.Consumer, error) {
	consumer, err := sarama.NewConsumer([]string{kafkaClient.kafkaConfig.KafkaUrl}, kafkaClient.saramaConfig)
	if err != nil {
		log.Printf("Error trying to connect to one of kafka brokers")
		return nil, err
	}
	kafkaClient.consumer = consumer
	return &consumer, nil
}

func (kafkaClient *KafkaClient) GetPartitions(topic string) (chan int32, error) {

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

func (kafkaClient *KafkaClient) ConsumePartition(topic string, partition int32) (sarama.PartitionConsumer, error) {
	partitionConsumer, err := kafkaClient.consumer.ConsumePartition(topic, partition, sarama.OffsetNewest)
	if err != nil {
		fmt.Printf("Error creating partition consumer for partition %d: %v\n", partition, err)
		return nil, err
	}
	return partitionConsumer, nil
}
