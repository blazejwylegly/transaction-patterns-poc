package listener

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/messaging"
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/transactions"
)

type FinalStepMessageHandler struct {
	txnService transactions.TxnService
}

func NewFinalStepMessageHandler(txnService transactions.TxnService) *FinalStepMessageHandler {
	return &FinalStepMessageHandler{txnService: txnService}
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

	handler.txnService.HandleFinalTxnStep(*txnContext, txnStep)
}
