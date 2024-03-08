package analytics

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
)

type SagaAnalyzer struct {
	client http.Client
}

func NewSagaAnalyzer() *SagaAnalyzer {
	return &SagaAnalyzer{client: http.Client{}}
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
