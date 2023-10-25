package main

import (
	"fmt"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/config"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/web"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

const configFileName = "config.yaml"

func main() {
	appConfig := config.New(configFileName)
	baseRouter := mux.NewRouter()

	web.InitDevApi(baseRouter, *appConfig)

	serverUrl := fmt.Sprintf("%s:%s", appConfig.Server.Host, appConfig.Server.Port)
	log.Fatal(http.ListenAndServe(serverUrl, baseRouter))
}
