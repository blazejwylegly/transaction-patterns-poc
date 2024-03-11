package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/analytics"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/config"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/db"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/messaging"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/messaging/listener"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/transactions"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/web"
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
	kafkaConsumer, _ := kafkaClient.NewConsumer()
	defer func(consumer sarama.Consumer) {
		err := consumer.Close()
		if err != nil {
			log.Fatalf("Error closing sarama kafka consumer: %v", err)
		}
	}(*kafkaConsumer)

	// DATABASE
	dbConnection := db.InitDbConnection(appConfig.GetDatabaseConfig())
	txRepository := db.NewDbRepository(dbConnection)

	// TX HANDLER
	txnService := transactions.NewTxnService(dbConnection)
	genericStepHandler := listener.NewGenericStepMessageHandler(*txnService)

	// CHOREOGRAPHY LISTENERS
	orderRequestListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.OrderPlaced)
	orderRequestListener.StartConsuming()

	itemsReservedListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.ItemsReserved)
	itemsReservedListener.StartConsuming()

	paymentProcessedListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.PaymentProcessed)
	paymentProcessedListener.StartConsuming()

	txnErrorListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.TxnError)
	txnErrorListener.StartConsuming()

	// ORCHESTRATION LISTENERS
	inventoryUpdateRequestListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.InventoryUpdateRequest)
	inventoryUpdateRequestListener.StartConsuming()

	inventoryUpdateStatusListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.InventoryUpdateStatus)
	inventoryUpdateStatusListener.StartConsuming()

	paymentRequestListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.PaymentRequest)
	paymentRequestListener.StartConsuming()

	paymentStatusListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.PaymentStatus)
	paymentStatusListener.StartConsuming()

	// FINAL STEP
	finalStepHandler := listener.NewFinalStepMessageHandler(*txnService)
	orderStatusListener := listener.NewTopicListener(*kafkaClient,
		finalStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.OrderStatus)
	orderStatusListener.StartConsuming()

	// PREPARE TESTING ENVIRONMENT
	txRepository.Purge()

	// START TESTING ROUTINE
	sagaAnalyzer := analytics.NewSagaAnalyzer()
	successful, failed := sagaAnalyzer.StartAnalysis()
	log.Printf("Success: %d, Failed: %d", successful, failed)

	// WEB
	router := mux.NewRouter()
	web.InitTxApi(router, txRepository)
	err := http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
