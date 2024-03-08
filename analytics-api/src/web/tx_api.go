package web

import (
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/db"
	"github.com/gorilla/mux"
	"net/http"
)

type TxApi struct {
	apiSubRouter *mux.Router
	txnRepo      *db.TxRepository
}

func InitTxApi(router *mux.Router, repository *db.TxRepository) *TxApi {
	devApiRouter := router.PathPrefix("/api/txn").Subrouter()
	api := &TxApi{
		apiSubRouter: devApiRouter,
		txnRepo:      repository,
	}
	api.initializeMappings()
	return api
}

func (api TxApi) initializeMappings() {
	api.apiSubRouter.HandleFunc("", api.handleGetTransactions()).Methods("GET")
}

func (api TxApi) handleGetTransactions() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		encoder := defaultEncoder(response)

		transactions := api.txnRepo.GetAll()
		err := encoder.Encode(transactions)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}
