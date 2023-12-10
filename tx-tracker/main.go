package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/config"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/db"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging/listener"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	// CONFIG
	appConfig := config.New(configFileName)

	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig.GetKafkaConfig())

	// KAFKA CONSUMER
	kafkaConsumer, err := kafkaClient.NewConsumer()
	defer func(consumer sarama.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka consumer: %v", err)
		}
	}(*kafkaConsumer)

	// TOPIC LISTENERS
	orderRequestListener := listener.NewListener(*kafkaClient, appConfig.GetKafkaConfig().KafkaTopics.OrderRequestsTopic)
	orderRequestListener.StartConsuming()

	itemsReservedListener := listener.NewListener(*kafkaClient, appConfig.GetKafkaConfig().KafkaTopics.ItemsReservedTopic)
	itemsReservedListener.StartConsuming()

	orderResultsListener := listener.NewListener(*kafkaClient, appConfig.GetKafkaConfig().KafkaTopics.OrderResultsTopic)
	orderResultsListener.StartConsuming()

	// DATABASE
	_ = db.InitDbConnection(appConfig.GetDatabaseConfig())

	// WEB
	router := mux.NewRouter()

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
