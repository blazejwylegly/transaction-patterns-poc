package web

import (
	"github.com/blazejwylegly/transactions-poc/products-service/config"
	"github.com/gorilla/mux"
	"net/http"
)

type DevApi struct {
	apiSubRouter *mux.Router
	config       config.Config
}

func InitDevApi(router *mux.Router, config config.Config) *DevApi {
	devApiRouter := router.PathPrefix("/dev/config").Subrouter()
	api := &DevApi{
		apiSubRouter: devApiRouter,
		config:       config,
	}
	api.initializeMappings()
	return api
}

func (api DevApi) initializeMappings() {
	api.apiSubRouter.HandleFunc("", api.handleGetConfig()).Methods("GET")
}

func (api DevApi) handleGetConfig() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		encoder := defaultEncoder(response)
		err := encoder.Encode(api.config)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}
