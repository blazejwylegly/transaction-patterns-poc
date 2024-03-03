package messaging

const (
	TransactionIdHeader        = "txn_id"
	TransactionNameHeader      = "txn_name"
	TransactionStartedAtHeader = "txn_started_at"
	StepIdHeader               = "txn_step_id"
	StepNameHeader             = "txn_step_name"
	StepExecutorHeader         = "txn_step_executor"
	StepStatusHeader           = "txn_step_status"
	StepStartedAtHeader        = "txn_step_started_at"
)

const (
	ItemReservationRequestedStep         = "ITEM_RESERVATION_REQUESTED"
	ItemReservationRollbackRequestedStep = "ITEM_RESERVATION_ROLLBACK_REQUESTED"
	StepStatusSuccess                    = "SUCCESS"
	StepStatusFailed                     = "FAILED"
)
