package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/google/uuid"
)

type Context struct {
	TransactionId        uuid.UUID
	TransactionName      string
	TransactionStartedAt string
	PreviousStepId       uuid.UUID
	PreviousStepName     string
	PreviousStepExecutor string
}

func (ctx *Context) ToHeaders() map[string]string {
	return map[string]string{
		messaging.TransactionIdHeader:        ctx.TransactionId.String(),
		messaging.TransactionNameHeader:      ctx.TransactionName,
		messaging.TransactionStartedAtHeader: ctx.TransactionStartedAt,
		messaging.StepIdHeader:               ctx.PreviousStepId.String(),
		messaging.StepNameHeader:             ctx.PreviousStepName,
		messaging.StepExecutorHeader:         ctx.PreviousStepExecutor,
	}
}

func ContextFromHeaders(headers map[string]string) (*Context, error) {
	txnIdHeader := headers[messaging.TransactionIdHeader]
	transactionId, err := uuid.Parse(txnIdHeader)
	if err != nil {
		return nil, err
	}
	stepIdHeader := headers[messaging.StepIdHeader]
	stepId, err := uuid.Parse(stepIdHeader)
	if err != nil {
		return nil, err
	}
	return &Context{
		TransactionId:        transactionId,
		TransactionName:      headers[messaging.TransactionNameHeader],
		TransactionStartedAt: headers[messaging.TransactionStartedAtHeader],
		PreviousStepId:       stepId,
		PreviousStepExecutor: headers[messaging.StepExecutorHeader],
		PreviousStepName:     headers[messaging.StepNameHeader],
	}, nil
}
