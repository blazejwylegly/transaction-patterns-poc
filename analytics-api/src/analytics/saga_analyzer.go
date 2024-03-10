package analytics

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

const createOrderUrl = "http://localhost:"

type SagaAnalyzer struct {
	client http.Client
}

func NewSagaAnalyzer() *SagaAnalyzer {
	return &SagaAnalyzer{client: http.Client{}}
}

func startAnalysis() {
	// create channel of requests
	requestsChannel := make(chan *http.Request)
	// create channel for responses
	initializeRequestPool(requestsChannel)
	//
}

func initializeRequestPool(channel chan *http.Request) {
	customerId := "0e97f795-5e09-4d55-862c-f087401eb62b"
	productId := "1acc1aa9-d40a-11ee-bb0e-80ce62448085"
	quantityOrdered := 1

	createOrderRequestBody := struct {
		CustomerId string `json:"customer_id"`
		OrderItems []struct {
			ProductId       string `json:"product_id"`
			QuantityOrdered int    `json:"quantity_ordered"`
		} `json:"order_items"`
	}{
		CustomerId: customerId,
		OrderItems: []struct {
			ProductId       string `json:"product_id"`
			QuantityOrdered int    `json:"quantity_ordered"`
		}{
			{
				ProductId:       productId,
				QuantityOrdered: quantityOrdered,
			},
		},
	}

	bodyJson, err := json.Marshal(createOrderRequestBody)
	if err != nil {
		log.Fatalf("Error trying to parse request object into string json")
		return
	}

	for i := 0; i < 100; i++ {
		request, err := http.NewRequest("POST", createOrderUrl, bodyJson)
	}
}

func (analyzer *SagaAnalyzer) sendCreateOrderRequest() {
	url := "http://example.com/api"

	// Define the JSON payload you want to send
	jsonStr := []byte(`{"key1":"value1","key2":"value2"}`)

	// Create a new HTTP request with the appropriate method, URL, and payload
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	if err != nil {
		fmt.Println("Error creating request:", err)
		return
	}

	// Set the Content-Type header to application/json
	req.Header.Set("Content-Type", "application/json")

	resp, err := analyzer.client.Do(req)
	if err != nil {
		fmt.Println("Error sending request:", err)
		return
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			fmt.Println("Error closing body  request:", err)
		}
	}(resp.Body)

	// Print the response status code and body
	fmt.Println("Response Status:", resp.Status)
	fmt.Println("Response Body:", resp.Body)
}
