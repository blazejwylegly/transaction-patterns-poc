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
		FlushFrequencyMs string `yaml:"flushFrequencyMs" envconfig:"KAFKA_FLUSH_FREQUENCY_MS"`
		OrdersTopic      string `yaml:"ordersTopic" envconfig:"KAFKA_ORDERS_TOPIC"`
	} `yaml:"kafka"`
}

type KafkaConfig struct {
	KafkaUrl              string
	KafkaFlushFrequencyMs string
	OrdersTopic           string
}

func (cfg *Config) GetDatabaseConfig() *KafkaConfig {
	return &KafkaConfig{
		KafkaUrl:              cfg.Kafka.Url,
		KafkaFlushFrequencyMs: cfg.Kafka.FlushFrequencyMs,
		OrdersTopic:           cfg.Kafka.OrdersTopic,
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
