package application

import (
	"errors"
	"fmt"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/database"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/events"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type OrderEventHandler struct {
	db *gorm.DB
}

func NewOrderEventHandler(db *gorm.DB) *OrderEventHandler {
	return &OrderEventHandler{db}
}

func (handler *OrderEventHandler) Handle(orderPlacedEvent events.OrderPlaced) (*events.OrderItemsReserved, error) {
	totalCost := 0.0
	for _, orderItem := range orderPlacedEvent.OrderItems {
		reservedProduct, err := handler.reserveProduct(orderItem.ProductId, orderItem.QuantityOrdered)
		if err != nil {
			return nil, err
		}
		totalCost += reservedProduct.Price * float64(orderItem.QuantityOrdered)
	}
	return &events.OrderItemsReserved{
		OrderID:    orderPlacedEvent.OrderID,
		CustomerID: orderPlacedEvent.CustomerID,
		TotalCost:  totalCost,
	}, nil
}

func (handler *OrderEventHandler) reserveProduct(productId uuid.UUID, requestedAmount int) (*database.Product, error) {
	var product *database.Product

	err := handler.db.Transaction(func(tx *gorm.DB) error {

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productId).Error; err != nil {
			log.Printf("Error reading product from db: %v", err)
			return err
		}

		if product == nil {
			log.Printf("Product with id %s not found!", productId.String())
			return nil
		}

		if product.Quantity < requestedAmount {
			log.Printf("Too many items requested for product with id %s!", productId.String())
			return errors.New(fmt.Sprintf("Too many items requested for product with id %s!", productId.String()))
		}

		if err := tx.Model(product).Update("quantity", product.Quantity-requestedAmount).Error; err != nil {
			log.Printf("Error updating product quantity: %v", err)
			return err
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return product, nil
}
