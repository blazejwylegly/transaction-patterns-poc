package application

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/blazejwylegly/transactions-poc/products-service/src/database"
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

func (handler *OrderEventHandler) Handle(orderPlacedEvent InventoryUpdateRequest) (*OrderItemsReserved, error) {
	tx := handler.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return nil, err
	}

	totalCost := 0.0
	for _, orderItem := range orderPlacedEvent.OrderItems {
		reservedProduct, err := handler.reserveProduct(tx, orderItem.ProductId, orderItem.QuantityOrdered)
		if err != nil {
			tx.Rollback()
			return nil, err
		}
		totalCost += reservedProduct.Price * float64(orderItem.QuantityOrdered)
	}
	err := handler.updateTransactionLog(tx, orderPlacedEvent)
	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	return &OrderItemsReserved{
		OrderID:    orderPlacedEvent.OrderID,
		CustomerID: orderPlacedEvent.CustomerID,
		TotalCost:  totalCost,
	}, nil
}

func (handler *OrderEventHandler) reserveProduct(tx *gorm.DB, productId uuid.UUID, requestedAmount int) (*database.Product, error) {
	var product *database.Product

	if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&product, productId).Error; err != nil {
		log.Printf("Error reading product from db: %v", err)
		return nil, err
	}

	if product == nil {
		log.Printf("Product with id %s not found!", productId.String())
		return nil, errors.New(fmt.Sprintf("Product with id %s not found!", productId.String()))
	}

	if product.Quantity < requestedAmount {
		log.Printf("Too many items requested for product with id %s!", productId.String())
		return nil, errors.New(fmt.Sprintf("Too few items for product with id %s!", productId.String()))
	}

	if err := tx.Model(product).Update("quantity", product.Quantity-requestedAmount).Error; err != nil {
		log.Printf("Error updating product quantity: %v", err)
		return nil, err
	}

	return product, nil
}

func (handler *OrderEventHandler) updateTransactionLog(tx *gorm.DB, event InventoryUpdateRequest) error {
	orderItemsJsonBody, err := json.Marshal(event.OrderItems)
	if err != nil {
		return err
	}

	txLog := &database.TxLogItem{
		TxLogItemId: uuid.New(),
		OrderID:     event.OrderID,
		OrderItems:  string(orderItemsJsonBody),
	}

	if err := tx.Save(txLog).Error; err != nil {
		log.Printf("Error saving txn log for orderId %s: %v", event.OrderID, err)
		return err
	}
	return nil
}

func (handler *OrderEventHandler) HandleRollback(orderFailedEvent OrderFailed) error {
	tx := handler.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if err := tx.Error; err != nil {
		return err
	}

	txLogItem, err := handler.restoreTransactionItem(tx, orderFailedEvent)

	if err != nil {
		tx.Rollback()
		return err
	}

	var orderItems []OrderItem
	_ = json.Unmarshal([]byte(txLogItem.OrderItems), &orderItems)

	for _, orderItem := range orderItems {
		err := handler.releaseOrderItem(tx, orderItem)
		if err != nil {
			tx.Rollback()
			return err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return err
	}
	return nil
}

func (handler *OrderEventHandler) restoreTransactionItem(tx *gorm.DB, event OrderFailed) (*database.TxLogItem, error) {
	var txLogItem *database.TxLogItem
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(&database.TxLogItem{OrderID: event.OrderID}).
		First(&txLogItem).
		Error
	if err != nil {
		log.Printf("FATAL ERROR - Cannot rollback transaction for orderId %s", event.OrderID)
		return nil, err
	}
	return txLogItem, nil
}

func (handler *OrderEventHandler) releaseOrderItem(tx *gorm.DB, item OrderItem) error {
	var product *database.Product
	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
		Where(&database.Product{ProductID: item.ProductId}).
		First(&product).
		Error
	if err != nil {
		log.Printf("Cannot rollback - Product with id %s not found!", item.ProductId)
		return err
	}

	err = tx.Model(product).Update("quantity", product.Quantity+item.QuantityOrdered).Error
	if err != nil {
		log.Printf("Cannot rollback - error trying to update quanitty of product with id %s", item.ProductId)
		return err
	}
	return nil
}
