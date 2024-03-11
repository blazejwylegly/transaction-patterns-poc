package listener

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/analytics"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/messaging"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/transactions"
	"log"
)

type FinalStepMessageHandler struct {
	txnService                     transactions.TxnService
	businessActionCompletedCounter analytics.TransactionCounter
	businessActionFailedCounter    analytics.TransactionCounter
	handledTransactionsCounter     analytics.TransactionCounter
	failedTransactionsCounter      analytics.TransactionCounter
}

func NewFinalStepMessageHandler(txnService transactions.TxnService) *FinalStepMessageHandler {
	return &FinalStepMessageHandler{
		txnService:                     txnService,
		businessActionCompletedCounter: analytics.TransactionCounter{},
		businessActionFailedCounter:    analytics.TransactionCounter{},
		handledTransactionsCounter:     analytics.TransactionCounter{},
		failedTransactionsCounter:      analytics.TransactionCounter{},
	}
}

func (handler *FinalStepMessageHandler) HandleTxnStepMessage(topic string, message sarama.ConsumerMessage) {
	headers := parseHeaders(message.Headers)
	txnContext, err := parseTransactionContext(headers)
	if err != nil {
		fmt.Printf("Error processing txn contetx for txnId %s and step %s\n",
			headers[messaging.StepNameHeader],
			headers[messaging.TransactionIdHeader])
		return
	}

	txnStep, err := parseTransactionStep(topic, headers)
	if err != nil {
		fmt.Printf("Error processing step message %s for txnId %s\n",
			headers[messaging.StepNameHeader],
			headers[messaging.TransactionIdHeader])
		return
	}

	transactionSuccessful, err := handler.txnService.HandleFinalTxnStep(*txnContext, txnStep)
	if err != nil {
		handler.failedTransactionsCounter.Inc()
	} else {
		handler.handledTransactionsCounter.Inc()
	}

	if transactionSuccessful {
		handler.businessActionCompletedCounter.Inc()
	} else {
		handler.businessActionFailedCounter.Inc()
	}

	log.Printf("Analysis completed! Successfully handled: %d, Failed to handle: %d",
		handler.handledTransactionsCounter.Get(),
		handler.failedTransactionsCounter.Get())
	log.Printf("%d of business actions completed with success, %d failed",
		handler.businessActionCompletedCounter.Get(),
		handler.businessActionFailedCounter.Get())
}
