package service

import (
	"errors"
	"fmt"
	"github.com/blazejwylegly/transactions-poc/products-service/src/order"
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"log"
)

type OrderRequestHandler struct {
	db *gorm.DB
}

func NewOrderRequestHandler(db *gorm.DB) OrderRequestHandler {
	return OrderRequestHandler{db}
}

func (handler OrderRequestHandler) Handle(request *orderModels.OrderRequestDto) {
	for _, orderItem := range request.OrderItems {
		handler.reserveProduct(orderItem.ProductId, orderItem.QuantityOrdered)
	}
}

func (handler OrderRequestHandler) reserveProduct(productId uuid.UUID, requestedAmount int) {

	err := handler.db.Transaction(func(tx *gorm.DB) error {
		var product *models.Product

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
		// todo: send rollback message
		return
	}

	// todo: send next saga message
}
