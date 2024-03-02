package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/application"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/application/saga"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/database"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging/listener"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/messaging/producer"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/web"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	// CONFIG
	appConfig := config.New(configFileName)

	if appConfig.ChoreographyModeEnabled() {
		log.Printf("Starting app in choreography mode")
		initChoreographyBasedApp(appConfig)
	} else {
		initOrchestrationBasedApp(appConfig)
	}

}

func initChoreographyBasedApp(appConfig *config.Config) {
	// WEB
	router := mux.NewRouter()

	web.InitPaymentsApi(router)
	web.InitDevApi(router, *appConfig)

	// DATABASE
	db := database.InitDbConnection(appConfig.GetDatabaseConfig())

	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig.GetKafkaConfig())

	// KAFKA PRODUCER
	kafkaProducer, err := kafkaClient.NewProducer()
	defer func(producer sarama.SyncProducer) {
		err := producer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka producer: %v", err)
		}
	}(*kafkaProducer)

	// KAFKA CONSUMER
	kafkaConsumer, err := kafkaClient.NewConsumer()
	defer func(consumer sarama.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka consumer: %v", err)
		}
	}(*kafkaConsumer)

	// SERVICE
	eventProducer := producer.NewSaramaProducer(*kafkaProducer)
	eventHandler := application.NewPaymentRequestedHandler(db)
	sagaCoordinator := saga.NewChoreographyCoordinator(*eventHandler, eventProducer, appConfig.GetKafkaConfig())

	paymentRequestListener := listener.NewListener(*kafkaClient,
		appConfig.GetKafkaConfig().KafkaTopics.ItemsReserved,
		sagaCoordinator)
	paymentRequestListener.StartConsuming()

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}

func initOrchestrationBasedApp(appConfig *config.Config) {

	// DATABASE
	db := database.InitDbConnection(appConfig.GetDatabaseConfig())

	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig.GetKafkaConfig())

	// KAFKA PRODUCER
	kafkaProducer, err := kafkaClient.NewProducer()
	defer func(producer sarama.SyncProducer) {
		err := producer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka producer: %v", err)
		}
	}(*kafkaProducer)

	// KAFKA CONSUMER
	kafkaConsumer, err := kafkaClient.NewConsumer()
	defer func(consumer sarama.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka consumer: %v", err)
		}
	}(*kafkaConsumer)

	// SERVICE
	eventProducer := producer.NewSaramaProducer(*kafkaProducer)
	eventHandler := application.NewPaymentRequestedHandler(db)
	sagaCoordinator := saga.NewOrchestrationCoordinator(*eventHandler, eventProducer, appConfig.GetKafkaConfig())

	paymentRequestListener := listener.NewListener(*kafkaClient,
		appConfig.GetKafkaConfig().KafkaTopics.PaymentRequested,
		sagaCoordinator)
	paymentRequestListener.StartConsuming()

	// WEB
	router := mux.NewRouter()

	web.InitPaymentsApi(router)
	web.InitDevApi(router, *appConfig)

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
