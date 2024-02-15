package handlers

import (
	"github.com/IBM/sarama"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/application/saga"
)

type OrderFailedHandler struct {
	sagaCoordinator saga.Coordinator
}

func NewOrderFailedHandler(sagaCoordinator saga.Coordinator) *OrderFailedHandler {
	return &OrderFailedHandler{sagaCoordinator: sagaCoordinator}
}

func (handler *OrderFailedHandler) Handle() func(chan *sarama.ConsumerMessage) {
	//TODO implement me
	panic("implement me")
}
