package listener

import (
	"encoding/json"
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/db"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/transactions"
	"github.com/google/uuid"
	"log"
	"sync"
	"time"
)

type TopicListener struct {
	kafkaClient    messaging.KafkaClient
	txnStepHandler transactions.TxnStepHandler
	topic          string
}

func NewTopicListener(kafkaClient messaging.KafkaClient,
	txnStepHandler transactions.TxnStepHandler,
	topic string) *TopicListener {
	return &TopicListener{kafkaClient: kafkaClient, txnStepHandler: txnStepHandler, topic: topic}
}

// StartConsuming fetches all available partitions from kafka broker, and runs a separate
// goroutine to process the messages from each partition.
func (listener *TopicListener) StartConsuming() {
	partitions, _ := listener.kafkaClient.GetPartitions(listener.topic)
	var wg sync.WaitGroup
	for partition := range partitions {
		go listener.consumePartition(&wg, partition)
	}
}

// consumePartition is responsible for consuming specified partition by running two goroutines.
// One of them is used to continuously fetch messages from kafka and the second one is continuously handling them.
func (listener *TopicListener) consumePartition(wg *sync.WaitGroup, partition int32) {
	messagesChannel := make(chan *sarama.ConsumerMessage)
	defer close(messagesChannel)

	partitionConsumer, err := listener.kafkaClient.ConsumePartition(listener.topic, partition)
	if err != nil {
		log.Printf("Error creating consumer for topic %s, partition %d: %v", listener.topic, partition, err)
	}

	wg.Add(1)
	go func(pc sarama.PartitionConsumer) {
		defer wg.Done()
		for message := range pc.Messages() {
			messagesChannel <- message
		}
	}(partitionConsumer)

	defer func(partitionConsumer sarama.PartitionConsumer) {
		err := partitionConsumer.Close()
		if err != nil {
			fmt.Printf("Error closing partition consumer for partition %s %v\n", partition, err)
		}
	}(partitionConsumer)

	go listener.processKafkaMessages(messagesChannel)
	select {}
}

func (listener *TopicListener) processKafkaMessages(messagesChannel chan *sarama.ConsumerMessage) {
	for msg := range messagesChannel {
		messagePayload := msg.Value
		headers := parseHeaders(msg.Headers)
		// TODO DLQ
		txnContext, err := parseTransactionContext(headers)
		if err != nil {
			fmt.Printf("Error processing txn contetx for txnId %s and step %s\n",
				headers[messaging.StepNameHeader],
				headers[messaging.TransactionIdHeader])
			return
		}

		txnStep, err := parseTransactionStep(messagePayload, headers)
		if err != nil {
			fmt.Printf("Error processing step message %s for txnId %s\n",
				headers[messaging.StepNameHeader],
				headers[messaging.TransactionIdHeader])
			return
		}

		listener.txnStepHandler.HandleTxnStep(*txnContext, txnStep)
	}
}

func parseTransactionContext(headers map[string]string) (*transactions.TxnContext, error) {
	txnIdHeader := headers[messaging.TransactionIdHeader]
	txnId, err := uuid.Parse(txnIdHeader)

	if err != nil {
		return nil, err
	}

	return &transactions.TxnContext{
		TxnId:        txnId,
		TxnName:      headers[messaging.TransactionNameHeader],
		TxnStartedAt: time.Now(),
	}, nil
}

func parseTransactionStep(messagePayload []byte, headers map[string]string) (*db.TransactionStep, error) {
	headersJson, err := json.Marshal(headers)
	if err != nil {
		return nil, err
	}

	stepId, err := uuid.Parse(headers[messaging.StepIdHeader])
	if err != nil {
		return nil, err
	}

	txnId, err := uuid.Parse(headers[messaging.TransactionIdHeader])
	if err != nil {
		return nil, err
	}

	txnStep := db.TransactionStep{}
	txnStep.StepId = stepId
	txnStep.TxnId = txnId
	txnStep.StepName = headers[messaging.StepNameHeader]
	txnStep.StepExecutor = headers[messaging.StepExecutorHeader]
	txnStep.Payload = string(messagePayload)
	txnStep.Headers = string(headersJson)
	return &txnStep, nil
}

func parseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}
