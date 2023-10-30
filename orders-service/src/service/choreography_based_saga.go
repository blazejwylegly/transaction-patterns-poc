package service

import (
	"github.com/blazejwylegly/transactions-poc/orders-service/src/messaging"
	"github.com/blazejwylegly/transactions-poc/orders-service/src/models"
	"log"
)

type OrderService struct {
	producer messaging.OrderRequestProducer
}

func NewOrderService(producer messaging.OrderRequestProducer) *OrderService {
	return &OrderService{
		producer: producer,
	}
}

func (svc *OrderService) BeginOrderPlacedTransaction(order *models.Order) {
	log.Printf("Trying to initiate txn with orderId %s", order.OrderID.String())
	svc.producer.Send(*order)
	log.Printf("Transaction with with orderId %s", order.OrderID.String())
}
