package listener

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/db"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/messaging"
	"github.com/blazejwylegly/transactions-poc/tx-tracker/src/transactions"
	"github.com/google/uuid"
	"time"
)

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
		TxnStatus:    headers[messaging.TransactionStatusHeader],
	}, nil
}

func parseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}

func parseTransactionStep(topic string, headers map[string]string) (*db.TransactionStep, error) {
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
	txnStep.Topic = topic
	txnStep.StepName = headers[messaging.StepNameHeader]
	txnStep.StepExecutor = headers[messaging.StepExecutorHeader]
	txnStep.StepStatus = headers[messaging.StepResultHeader]
	txnStep.StepStartedAt = headers[messaging.StepStartedAtHeader]
	return &txnStep, nil
}
