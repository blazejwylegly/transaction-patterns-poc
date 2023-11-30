package web

import (
	"github.com/gorilla/mux"
	"net/http"
)

type PaymentsApi struct {
	apiSubRouter *mux.Router
}

func InitPaymentsApi(router *mux.Router) *PaymentsApi {
	paymentsApiRouter := router.PathPrefix("/api/payments").Subrouter()
	api := &PaymentsApi{
		apiSubRouter: paymentsApiRouter,
	}
	api.initializeMappings()
	return api
}

func (api *PaymentsApi) initializeMappings() {
	api.apiSubRouter.HandleFunc("/{orderId}", api.handleFindByOrderId()).Methods("GET")
	api.apiSubRouter.HandleFunc("", api.handleFindAllPayments()).Methods("GET")
}

func (api *PaymentsApi) handleFindAllPayments() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		encoder := defaultEncoder(response)
		err := encoder.Encode("")
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (api *PaymentsApi) handleFindByOrderId() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		encoder := defaultEncoder(response)
		err := encoder.Encode("")
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}
