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
	StepStartedAtHeader        = "txn_step_started_at"
)

const (
	ItemReservationRequestedStep         = "ITEM_RESERVATION_REQUESTED"
	ItemReservationRollbackRequestedStep = "ITEM_RESERVATION_ROLLBACK_REQUESTED"
	PaymentRequestedStep                 = "PAYMENT_REQUESTED"
	OrderCompletedStep                   = "ORDER_COMPLETED"
	OrderCancelledStep                   = "ORDER_CANCELLED"
	StepStatusPending                    = "PENDING"
	StepStatusSuccess                    = "SUCCESS"
	StepStatusFailed                     = "FAILED"
	TxnStatusPending                     = "PENDING"
	TxnStatusRollbackInProgress          = "ROLLBACK_IN_PROGRESS"
	TxnStatusSuccess                     = "SUCCESS"
	TxnStatusFailed                      = "FAILED"
)

func ParseHeaders(recordHeaders []*sarama.RecordHeader) map[string]string {
	headers := make(map[string]string, len(recordHeaders))
	for _, recordHeader := range recordHeaders {
		headers[string(recordHeader.Key)] = string(recordHeader.Value)
	}
	return headers
}
