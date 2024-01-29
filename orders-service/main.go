package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/database"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging/listener"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/saga"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/web"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	appConfig := config.New(configFileName)

	if appConfig.ChoreographyModeEnabled() {
		initChoreographyBasedApp(*appConfig)
	} else if appConfig.OrchestrationModeEnabled() {
		initOrchestrationBasedApp(*appConfig)
	}
}

func initChoreographyBasedApp(appConfig config.Config) {

	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig)
	saramaProducer := kafkaClient.NewProducer()

	defer func() {
		err := saramaProducer.Close()
		if err != nil {
			log.Printf("Error closing kafka producer %v", err)
		}
	}()
	orderProducer := messaging.NewSaramaProducer(saramaProducer)

	// SERVICE
	choreographyOverseer := saga.NewChoreographyCoordinator(orderProducer, appConfig.GetKafkaConfig())

	// WEB
	baseRouter := mux.NewRouter()
	web.InitDevApi(baseRouter, appConfig)
	web.NewOrderApi(baseRouter, choreographyOverseer)

	// APP SERVER
	log.Fatal(http.ListenAndServe(appConfig.GetServerUrl(), baseRouter))
}

func initOrchestrationBasedApp(appConfig config.Config) {
	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig)

	// KAFKA PRODUCER
	saramaProducer := kafkaClient.NewProducer()

	defer func() {
		err := saramaProducer.Close()
		if err != nil {
			log.Printf("Error closing kafka producer %v", err)
		}
	}()

	orderProducer := messaging.NewSaramaProducer(saramaProducer)

	// KAFKA CONSUMER
	kafkaConsumer, _ := kafkaClient.NewConsumer()
	defer func(consumer sarama.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka consumer: %v", err)
		}
	}(*kafkaConsumer)

	// DATABASE
	dbConnection := database.InitDbConnection(appConfig.GetDatabaseConfig())
	sagaRepository := database.NewSagaRepository(dbConnection)

	// SERVICE
	sagaLogger := saga.NewLogger(sagaRepository)
	orchestrator := saga.NewOrchestrationCoordinator(*sagaLogger, orderProducer, appConfig.GetKafkaConfig())

	// TOPIC LISTENERS
	orderRequestListener := listener.NewTopicListener(*kafkaClient,
		orchestrator,
		appConfig.GetKafkaConfig().KafkaTopics.ItemsReservedTopic)
	orderRequestListener.StartConsuming()

	orderResultListener := listener.NewTopicListener(*kafkaClient,
		orchestrator
		appConfig.GetKafkaConfig().KafkaTopics.OrderResultsTopic)
	orderResultListener.StartConsuming()

	// WEB
	baseRouter := mux.NewRouter()
	web.InitDevApi(baseRouter, appConfig)
	web.NewOrderApi(baseRouter, orchestrator)

	// APP SERVER
	log.Fatal(http.ListenAndServe(appConfig.GetServerUrl(), baseRouter))
}
