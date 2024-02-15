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
	txHandler := transactions.NewTxnStepHandler(dbConnection)

	// TOPIC LISTENERS
	orderRequestListener := listener.NewTopicListener(*kafkaClient,
		*txHandler,
		appConfig.GetKafkaConfig().KafkaTopics.OrderPlaced)
	orderRequestListener.StartConsuming()

	itemsReservedListener := listener.NewTopicListener(*kafkaClient,
		*txHandler,
		appConfig.GetKafkaConfig().KafkaTopics.ItemsReserved)
	itemsReservedListener.StartConsuming()

	paymentCompletedListener := listener.NewTopicListener(*kafkaClient,
		*txHandler,
		appConfig.GetKafkaConfig().KafkaTopics.PaymentProcessed)
	paymentCompletedListener.StartConsuming()

	orderResultsListener := listener.NewTopicListener(*kafkaClient,
		*txHandler,
		appConfig.GetKafkaConfig().KafkaTopics.TxnError)
	orderResultsListener.StartConsuming()

	orderFailedListener := listener.NewTopicListener(*kafkaClient,
		*txHandler,
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
