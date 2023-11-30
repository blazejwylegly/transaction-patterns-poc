package main

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"github.com/blazejwylegly/transactions-poc/payments-service/src/web"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	// CONFIG
	appConfig := config.New(configFileName)

	// WEB
	router := mux.NewRouter()

	web.InitPaymentsApi(router)
	web.InitDevApi(router, *appConfig)

	err := http.ListenAndServe(appConfig.GetServerUrl(), router)
	if err != nil {
		log.Fatal(err)
	}
}
