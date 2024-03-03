package main

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/config"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/db"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging/listener"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/transactions"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/web"
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

	// DATABASE
	dbConnection := db.InitDbConnection(appConfig.GetDatabaseConfig())
	txRepository := db.NewDbRepository(dbConnection)

	// TX HANDLER
	txnService := transactions.NewTxnService(dbConnection)
	genericStepHandler := listener.NewGenericStepMessageHandler(*txnService)

	// TOPIC LISTENERS
	orderRequestListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.OrderPlaced)
	orderRequestListener.StartConsuming()

	itemsReservedListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.ItemsReserved)
	itemsReservedListener.StartConsuming()

	paymentCompletedListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.PaymentProcessed)
	paymentCompletedListener.StartConsuming()

	orderResultsListener := listener.NewTopicListener(*kafkaClient,
		genericStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.TxnError)
	orderResultsListener.StartConsuming()

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
	orderFailedListener := listener.NewTopicListener(*kafkaClient,
		finalStepHandler.HandleTxnStepMessage,
		appConfig.GetKafkaConfig().KafkaTopics.OrderStatus)
	orderFailedListener.StartConsuming()

	// WEB
	router := mux.NewRouter()
	web.InitTxApi(router, txRepository)

	err = http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
