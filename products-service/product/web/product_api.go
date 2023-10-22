package web

import (
	"encoding/json"
	"github.com/blazejwylegly/transactions-poc/products-service/product/models"
	"github.com/blazejwylegly/transactions-poc/products-service/product/service"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

type ProductApi struct {
	apiSubRouter   *mux.Router
	productService *service.ProductService
}

func InitProductApi(router *mux.Router, productService *service.ProductService) *ProductApi {
	productApiRouter := router.PathPrefix("/api/product").Subrouter()
	api := &ProductApi{
		apiSubRouter:   productApiRouter,
		productService: productService,
	}
	api.initializeMappings()
	return api
}

func (api ProductApi) initializeMappings() {
	api.apiSubRouter.HandleFunc("/{productId}", api.handleFindByProductId()).Methods("GET")
	api.apiSubRouter.HandleFunc("/", api.handleFindAllProducts()).Methods("GET")

	api.apiSubRouter.HandleFunc("/", api.handleSaveProduct()).Methods("POST")
	api.apiSubRouter.HandleFunc("/all", api.handleSaveAllProducts()).Methods("POST")

	api.apiSubRouter.HandleFunc("/", api.handleUpdateInventory()).Methods("PUT")

	api.apiSubRouter.HandleFunc("/{productId}/quantity", api.handleUpdateProductQuantity()).Methods("PATCH")
}

func (api ProductApi) handleUpdateInventory() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		product, err := readObject(request, &models.ProductDto{})
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		}
		api.productService.SaveOrUpdate(product)
	}
}

func (api ProductApi) handleFindAllProducts() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		encoder := defaultEncoder(response)
		products := api.productService.FindAll()

		err := encoder.Encode(products)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (api ProductApi) handleFindByProductId() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		productId, err := strconv.Atoi(mux.Vars(request)["productId"])
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		}
		product := api.productService.FindById(productId)
		encoder := defaultEncoder(response)
		err = encoder.Encode(product)
		if err != nil {
			http.Error(response, err.Error(), http.StatusInternalServerError)
		}
	}
}

func (api ProductApi) handleSaveProduct() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		product, err := readObject(request, &models.ProductDto{})
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		}
		api.productService.SaveOrUpdate(product)
	}
}

func (api ProductApi) handleSaveAllProducts() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		products, err := readManyObjects(request)
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		}
		api.productService.SaveAll(products)
	}
}

func (api ProductApi) handleUpdateProductQuantity() func(http.ResponseWriter, *http.Request) {
	return func(response http.ResponseWriter, request *http.Request) {
		productId, err := strconv.Atoi(mux.Vars(request)["productId"])
		if err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		}

		var requestBody struct {
			Quantity int `json:"quantity"`
		}

		if err := json.NewDecoder(request.Body).Decode(&requestBody); err != nil {
			http.Error(response, err.Error(), http.StatusBadRequest)
		}

		api.productService.UpdateQuantity(productId, requestBody)
	}
}
