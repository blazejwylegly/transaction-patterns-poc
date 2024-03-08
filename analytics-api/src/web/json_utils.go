package web

import (
	"encoding/json"
	"net/http"
)

func defaultEncoder(writer http.ResponseWriter) *json.Encoder {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "\t")
	return encoder
}
