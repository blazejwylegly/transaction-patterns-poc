package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application"
	"github.com/blazejwylegly/transactions-poc/products-service/src/application/saga"
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/database"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/products-service/src/messaging/handlers"
	"github.com/blazejwylegly/transactions-poc/products-service/src/web"

	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	// CONFIG
	appConfig := config.New(configFileName)

	if appConfig.ChoreographyModeEnabled() {
		initChoreographyBasedApp(appConfig)
	} else {
		initOrchestrationBasedApp(appConfig)
	}
}

func initOrchestrationBasedApp(appConfig *config.Config) {
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
	eventProducer := messaging.NewSaramaProducer(*kafkaProducer)

	productService := application.NewProductService(productRepo)
	eventHandler := application.NewOrderEventHandler(db)

	sagaCoordinator := saga.NewOrchestrationCoordinator(*eventHandler, eventProducer, appConfig.GetKafkaConfig())

	reservationRequestHandler := handlers.NewItemReservationRequestHandler(sagaCoordinator)
	reservationRequestListener := messaging.NewListener(*kafkaClient,
		reservationRequestHandler.Handle,
		appConfig.Kafka.Topics.InventoryUpdateRequest)
	reservationRequestListener.StartConsuming()

	// WEB
	router := mux.NewRouter()

	web.InitProductApi(router, productService)
	web.InitDevApi(router, *appConfig)

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}

func initChoreographyBasedApp(appConfig *config.Config) {
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
	eventProducer := messaging.NewSaramaProducer(*kafkaProducer)

	productService := application.NewProductService(productRepo)
	eventHandler := application.NewOrderEventHandler(db)

	sagaCoordinator := saga.NewChoreographyCoordinator(*eventHandler, eventProducer, appConfig.GetKafkaConfig())

	itemReservationRequestHandler := handlers.NewItemReservationRequestHandler(sagaCoordinator)
	itemReservationRequestListener := messaging.NewListener(*kafkaClient,
		itemReservationRequestHandler.Handle,
		appConfig.Kafka.Topics.InventoryUpdateRequest)
	itemReservationRequestListener.StartConsuming()

	orderFailedHandler := handlers.NewOrderFailedHandler(sagaCoordinator)
	orderFailedListener := messaging.NewListener(*kafkaClient,
		orderFailedHandler.Handle,
		appConfig.Kafka.Topics.TxnError)
	orderFailedListener.StartConsuming()

	// WEB
	router := mux.NewRouter()

	web.InitProductApi(router, productService)
	web.InitDevApi(router, *appConfig)

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
