package web

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/service"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type OrderApi struct {
	router       *mux.Router
	orderService *service.OrderService
}

func InitOrderRouting(router *mux.Router, orderService *service.OrderService) *OrderApi {
	ordersRouter := router.PathPrefix("/order").Subrouter()
	api := &OrderApi{
		router:       ordersRouter,
		orderService: orderService,
	}
	api.initializeMappings()
	return api
}

func (api OrderApi) initializeMappings() {
	api.router.HandleFunc("", api.handlePlaceOrderMapping()).Methods("POST")
}

func (api OrderApi) handlePlaceOrderMapping() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {

		order := models.NewOrder()
		err := readOrder(request, order)
		if err != nil {
			log.Fatal("Error creating new order for customer!")
			return
		}

		api.orderService.BeginOrderPlacedTransaction(order)

		err = defaultEncoder(writer).Encode(order)
		if err != nil {
			log.Fatal("Error encoding created order!")
			return
		}
	}
}
