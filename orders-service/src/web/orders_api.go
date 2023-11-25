package web

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/saga"
	"github.com/gorilla/mux"
	"log"
	"net/http"
)

type OrderApi struct {
	router          *mux.Router
	sagaCoordinator saga.Coordinator
}

func NewOrderApi(router *mux.Router, coordinator saga.Coordinator) *OrderApi {
	ordersRouter := router.PathPrefix("/order").Subrouter()
	api := &OrderApi{
		router:          ordersRouter,
		sagaCoordinator: coordinator,
	}
	api.initializeMappings()
	return api
}

func (api *OrderApi) initializeMappings() {
	api.router.HandleFunc("", api.handlePlaceOrderMapping()).Methods("POST")
}

func (api *OrderApi) handlePlaceOrderMapping() func(writer http.ResponseWriter, request *http.Request) {
	return func(writer http.ResponseWriter, request *http.Request) {

		order := models.NewOrder()
		err := readOrder(request, order)
		if err != nil {
			log.Fatal("Error creating new order for customer!")
			return
		}

		api.sagaCoordinator.BeginOrderPlacedTransaction(*order)

		err = defaultEncoder(writer).Encode(order)
		if err != nil {
			log.Fatal("Error encoding created order!")
			return
		}
	}
}
