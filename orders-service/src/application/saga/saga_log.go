package saga

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/database"
)

// TODO
type Logger struct {
	repository database.SagaRepository
}

func (logger *Logger) registerNewSaga() {

}

func (logger *Logger) registerSagaStep() {

}

func (logger *Logger) terminateSaga() {

}

func (logger *Logger) finalizeSaga() {

}

func NewLogger(sagaRepository *database.SagaRepository) *Logger {
	return &Logger{*sagaRepository}
}
