package web

import (
	"encoding/json"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/application"
	"net/http"
)

func readOrder(request *http.Request, order *application.Order) error {
	err := json.NewDecoder(request.Body).Decode(order)
	if err != nil {
		return err
	}
	return nil
}

func defaultEncoder(writer http.ResponseWriter) *json.Encoder {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	return encoder
}
