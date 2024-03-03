package listener

import (
	"fmt"
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/transactions"
)

type GenericStepMessageHandler struct {
	txnService transactions.TxnService
}

func NewGenericStepMessageHandler(txnService transactions.TxnService) *GenericStepMessageHandler {
	return &GenericStepMessageHandler{txnService: txnService}
}

func (handler *GenericStepMessageHandler) HandleTxnStepMessage(topic string, message sarama.ConsumerMessage) {
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

	handler.txnService.HandleTxnStep(*txnContext, txnStep)
}
