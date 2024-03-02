package config

import (
	"fmt"
	"github.com/kelseyhightower/envconfig"
	"gopkg.in/yaml.v3"
	"log"
	"os"
)

type ApplicationMode int

const (
	Choreography  ApplicationMode = 0
	Orchestration ApplicationMode = 1
)

type Config struct {
	Application struct {
		Mode ApplicationMode `yaml:"mode" envconfig:"APPLICATION_MODE"`
	}

	Server struct {
		Host string `yaml:"host" envconfig:"SERVER_HOST"`
		Port string `yaml:"port" envconfig:"SERVER_PORT"`
	} `yaml:"server"`

	Database struct {
		DbHost     string `yaml:"host" envconfig:"DB_HOST"`
		DbPort     string `yaml:"port" envconfig:"DB_PORT"`
		DbUsername string `yaml:"username" envconfig:"DB_USERNAME"`
		DbPassword string `yaml:"password" envconfig:"DB_PASSWORD"`
		DbName     string `yaml:"name" envconfig:"DB_NAME"`
	} `yaml:"database"`

	Kafka struct {
		Url              string `yaml:"url" envconfig:"KAFKA_URL"`
		FlushFrequencyMs int32  `yaml:"flushFrequencyMs" envconfig:"KAFKA_FLUSH_FREQUENCY_MS"`
		Topics           struct {
			ItemsReserved    string `yaml:"itemsReserved" envconfig:"KAFKA_ITEMS_RESERVED_TOPIC"`
			PaymentProcessed string `yaml:"paymentProcessed" envconfig:"KAFKA_PAYMENT_PROCESSED_TOPIC"`
			TxnError         string `yaml:"txnError" envconfig:"KAFKA_TXN_ERROR_TOPIC"`
			PaymentRequested string `yaml:"paymentRequested" envconfig:"KAFKA_PAYMENT_REQUESTED_TOPIC"`
			PaymentStatus    string `yaml:"paymentStatus" envconfig:"KAFKA_PAYMENT_PROCESSED_TOPIC"`
		} `yaml:"topics"`
	} `yaml:"kafka"`
}

type KafkaConfig struct {
	KafkaUrl              string
	KafkaFlushFrequencyMs int32
	KafkaTopics           KafkaTopics
}

type KafkaTopics struct {
	ItemsReserved    string
	PaymentProcessed string
	TxnError         string
	PaymentRequested string
	PaymentStatus    string
}

type DatabaseConfig struct {
	DbHost     string
	DbPort     string
	DbUsername string
	DbPassword string
	DbName     string
}

func (cfg *Config) GetServerUrl() string {
	return fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
}

func (dbConfig *DatabaseConfig) GetDbUrl() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbConfig.DbUsername,
		dbConfig.DbPassword,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbName,
	)
}

func (cfg *Config) GetKafkaConfig() KafkaConfig {
	return KafkaConfig{
		KafkaUrl:              cfg.Kafka.Url,
		KafkaFlushFrequencyMs: cfg.Kafka.FlushFrequencyMs,
		KafkaTopics: KafkaTopics{
			ItemsReserved:    cfg.Kafka.Topics.ItemsReserved,
			PaymentProcessed: cfg.Kafka.Topics.PaymentProcessed,
			TxnError:         cfg.Kafka.Topics.TxnError,
			PaymentRequested: cfg.Kafka.Topics.PaymentRequested,
			PaymentStatus:    cfg.Kafka.Topics.PaymentStatus,
		},
	}
}

func (cfg *Config) GetDatabaseConfig() DatabaseConfig {
	return DatabaseConfig{
		DbHost:     cfg.Database.DbHost,
		DbPort:     cfg.Database.DbPort,
		DbUsername: cfg.Database.DbUsername,
		DbPassword: cfg.Database.DbPassword,
		DbName:     cfg.Database.DbName,
	}
}

func (cfg *Config) ChoreographyModeEnabled() bool {
	return cfg.Application.Mode == Choreography
}

func (cfg *Config) OrchestrationModeEnabled() bool {
	return cfg.Application.Mode == Orchestration
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

	readEnv(&config)
	return &config
}

func readEnv(cfg *Config) {
	err := envconfig.Process("", cfg)
	if err != nil {
		os.Exit(2)
	}
}
