package main

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/service"
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
	kafkaClient := *messaging.NewKafkaClient(appConfig)
	kafkaProducer := kafkaClient.NewProducer()

	defer func() {
		err := kafkaProducer.Close()
		if err != nil {
			log.Printf("Error closing kafka producer %v", err)
		}
	}()
	orderRequestProducer := *messaging.NewProducer(kafkaProducer, appConfig)

	// SERVICE
	choreographyOrderService := service.NewOrderService(orderRequestProducer)

	// WEB
	baseRouter := mux.NewRouter()
	web.InitDevApi(baseRouter, appConfig)
	web.InitOrderRouting(baseRouter, choreographyOrderService)

	// APP SERVER
	log.Fatal(http.ListenAndServe(appConfig.GetServerUrl(), baseRouter))
}
