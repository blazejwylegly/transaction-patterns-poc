package analytics

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"
	"sync"
)

const createOrderUrl = "http://localhost:8080/order"

type SagaAnalyzer struct {
	client                    http.Client
	successfulRequestsCounter TransactionCounter
	failedRequestsCounter     TransactionCounter
}

func NewSagaAnalyzer() *SagaAnalyzer {
	return &SagaAnalyzer{
		client:                    http.Client{},
		successfulRequestsCounter: TransactionCounter{},
		failedRequestsCounter:     TransactionCounter{},
	}
}

func (analyzer *SagaAnalyzer) StartAnalysis() (int32, int32) {
	// create channel of requests
	requestsChannel := make(chan *http.Request, 100)
	analyzer.initializeRequestPool(requestsChannel)

	wg := &sync.WaitGroup{}
	// Add lock for worker pool
	wg.Add(1)
	go func() {
		// Wait until worker pool finishes
		defer wg.Done()
		// Start worker pool
		analyzer.startWorkerPool(requestsChannel)
	}()

	// Wait for the worker pool goroutine to finish
	wg.Wait()
	log.Print("Finished triggering requests")
	return analyzer.successfulRequestsCounter.Get(), analyzer.failedRequestsCounter.Get()
}

func (analyzer *SagaAnalyzer) startWorkerPool(requests chan *http.Request) {
	maxConcurrency := 10
	workerWg := &sync.WaitGroup{}
	for i := 0; i < maxConcurrency; i++ {
		// Add worker to group
		workerWg.Add(1)
		// Concurrently...
		go func() {
			// Wait with removing worker from group until he finishes his work
			defer workerWg.Done()
			// ... start worker
			analyzer.startWorker(requests)
		}()
	}
	// Wait for all workers
	workerWg.Wait()
	log.Print("Worker pool terminating")
}

func (analyzer *SagaAnalyzer) startWorker(requests chan *http.Request) {
	for request := range requests {
		_, err := analyzer.client.Do(request)
		if err != nil {
			log.Printf("Error executing request %v", err)
			analyzer.failedRequestsCounter.Inc()
		} else {
			log.Print("Request executed successfully")
			analyzer.successfulRequestsCounter.Inc()
		}
	}
	log.Print("Worker terminating")
}

func (analyzer *SagaAnalyzer) initializeRequestPool(channel chan *http.Request) {
	customerId := "0e97f795-5e09-4d55-862c-f087401eb62b"
	productId := "ceb9933b-7ff7-11ee-9e27-2cf05d892018"
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

	body, err := json.Marshal(createOrderRequestBody)
	if err != nil {
		log.Fatalf("Error trying to parse request object into string json")
		return
	}

	for i := 0; i < 100; i++ {
		request, err := http.NewRequest("POST", createOrderUrl, bytes.NewBuffer(body))
		if err != nil {
			log.Printf("Error creating request")
		}
		channel <- request
	}

	log.Print("Finished populating requests channel")
	close(channel)
}
