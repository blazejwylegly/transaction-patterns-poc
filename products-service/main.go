package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/database"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging/listener"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging/producer"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/saga"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/web"

	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	// CONFIG
	appConfig := config.New(configFileName)

	// DB
	db := database.InitDbConnection(appConfig.GetDatabaseConfig())
	productRepo := database.NewProductRepository(db)

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
	productService := application.NewProductService(productRepo)
	eventHandler := application.NewOrderEventHandler(db)
	sagaCoordinator := saga.NewCoordinator(*eventHandler, eventProducer, appConfig.GetKafkaConfig())

	orderListener := listener.NewListener(*kafkaClient, appConfig.GetKafkaConfig(), *sagaCoordinator)
	orderListener.StartConsuming()

	// WEB
	router := mux.NewRouter()

	web.InitProductApi(router, productService)
	web.InitDevApi(router, *appConfig)

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
