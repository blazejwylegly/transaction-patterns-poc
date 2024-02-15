package messaging

import "github.com/IBM/sarama"

const (
	TransactionIdHeader        = "txn_id"
	TransactionNameHeader      = "txn_name"
	TransactionStartedAtHeader = "txn_started_at"
	TransactionStatusHeader    = "txn_status"
	StepIdHeader               = "txn_step_id"
	StepNameHeader             = "txn_step_name"
	StepExecutorHeader         = "txn_step_executor"
	StepStatusHeader           = "txn_step_status"
)

func ParseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}
