package main

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/saga"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/web"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {

	// CONFIG
	appConfig := *config.New(configFileName)

	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig)
	saramaProducer := kafkaClient.NewProducer()

	defer func() {
		err := saramaProducer.Close()
		if err != nil {
			log.Printf("Error closing kafka producer %v", err)
		}
	}()
	orderProducer := messaging.NewSaramaProducer(saramaProducer, appConfig.GetKafkaConfig())

	// SERVICE
	choreographyOrderService := saga.NewCoordinator(orderProducer, appConfig.GetKafkaConfig())

	// WEB
	baseRouter := mux.NewRouter()
	web.InitDevApi(baseRouter, appConfig)
	web.NewOrderApi(baseRouter, *choreographyOrderService)

	// APP SERVER
	log.Fatal(http.ListenAndServe(appConfig.GetServerUrl(), baseRouter))
}
