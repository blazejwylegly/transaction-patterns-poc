package messaging

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"log"
	"os"
	"time"
)

type KafkaClient struct {
	saramaConfig *sarama.Config
	kafkaConfig  *config.KafkaConfig
}

func NewKafkaClient(cfg config.Config) *KafkaClient {

	saramaConfig := sarama.NewConfig()
	sarama.Logger = log.New(os.Stdout, "[sarama] ", log.LstdFlags)
	saramaConfig.Producer.Flush.Frequency = time.Duration(cfg.GetKafkaConfig().KafkaFlushFrequencyMs) * time.Millisecond
	saramaConfig.Producer.Compression = sarama.CompressionGZIP
	saramaConfig.Producer.RequiredAcks = sarama.WaitForLocal
	saramaConfig.Producer.Return.Errors = true
	saramaConfig.Producer.Return.Successes = true
	return &KafkaClient{
		saramaConfig: saramaConfig,
		kafkaConfig:  cfg.GetKafkaConfig(),
	}
}

// NewProducer exposes new instance of a synchronous producer.
// Synchronous producers are used to ensure that the order will be eventually processed when submitted.
func (kafkaClient *KafkaClient) NewProducer() sarama.SyncProducer {
	producer, err := sarama.NewSyncProducer([]string{kafkaClient.kafkaConfig.KafkaUrl}, kafkaClient.saramaConfig)
	if err != nil {
		log.Printf("Error creating producer: %v", err)
	}
	return producer
}
