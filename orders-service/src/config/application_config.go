package config

import (
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type Config struct {
	Server struct {
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
	} `yaml:"server"`

	Kafka struct {
		Url              string `yaml:"url" envconfig:"KAFKA_URL"`
		FlushFrequencyMs int32  `yaml:"flushFrequencyMs" envconfig:"KAFKA_FLUSH_FREQUENCY_MS"`
		Topics           struct {
			OrderRequestsTopic string `yaml:"orderRequestsTopic" envconfig:"KAFKA_ORDERS_TOPIC"`
		} `yaml:"topics"`
	} `yaml:"kafka"`
}

type KafkaConfig struct {
	KafkaUrl              string
	KafkaFlushFrequencyMs int32
}

type KafkaTopics struct {
	OrderRequestsTopic string
}

func (cfg *Config) GetKafkaConfig() *KafkaConfig {
	return &KafkaConfig{
		KafkaUrl:              cfg.Kafka.Url,
		KafkaFlushFrequencyMs: cfg.Kafka.FlushFrequencyMs,
	}
}

func (cfg *Config) GetKafkaTopics() *KafkaTopics {
	return &KafkaTopics{
		OrderRequestsTopic: cfg.Kafka.Topics.OrderRequestsTopic,
	}
}

func New(configFileName string) *Config {
	var config Config
	file, err := os.Open(configFileName)
	if err != nil {
		os.Exit(2)
	}

	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(file)

	decoder := yaml.NewDecoder(file)
	err = decoder.Decode(&config)
	if err != nil {
		os.Exit(2)
	}

	readEnvironmentVariables(&config)
	return &config
}

func readEnvironmentVariables(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		os.Exit(2)
	}
}
