package web

import (
	"encoding/json"
	"github.com/blazejwylegly/transactions-poc/products-service/src/dto"
	"io"
	"net/http"
)

func readObject[T any](request *http.Request, object *T) (*T, error) {
	err := json.NewDecoder(request.Body).Decode(object)
	if err != nil {
		return nil, err
	}
	return object, nil
}

func readManyObjects(request *http.Request) ([]*dto.ProductDto, error) {
	bytes, err := io.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	var products []*dto.ProductDto

	err = json.Unmarshal(bytes, &products)
	if err != nil {
		return nil, err
	}
	return products, nil
}

func defaultEncoder(writer http.ResponseWriter) *json.Encoder {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	return encoder
}
