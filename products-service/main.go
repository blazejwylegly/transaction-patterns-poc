package main

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/config"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/data"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/messaging"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/service"
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
	productRepo := data.NewProductRepository(appConfig.GetDatabaseConfig())
	productService := service.NewProductService(&productRepo)

	// KAFKA
	kafkaClient := messaging.NewKafkaClient(appConfig.GetKafkaConfig())
	orderListener := messaging.NewListener(kafkaClient, appConfig.GetKafkaConfig())
	orderListener.StartConsuming()
	// WEB
	router := mux.NewRouter()

	web.InitProductApi(router, productService)
	web.InitDevApi(router, *appConfig)

	err := http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
