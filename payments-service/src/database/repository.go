package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type PaymentRepository interface {
	FindAll() []*Payment
	SaveOrUpdate(*Payment)
	FindByOrderId(uuid.UUID) *Payment
}

type PostgresPaymentRepository struct {
	db *gorm.DB
}

func (repo *PostgresPaymentRepository) FindAll() []*Payment {
	var payments []*Payment
	repo.db.Find(&payments)
	return payments
}

func (repo *PostgresPaymentRepository) FindByOrderId(orderId uuid.UUID) *Payment {
	var payment *Payment
	repo.db.First(&payment, orderId)
	return payment
}

func (repo *PostgresPaymentRepository) SaveOrUpdate(payment *Payment) {
	repo.db.Save(payment)
}

func NewPaymentRepository(db *gorm.DB) *PostgresPaymentRepository {
	return &PostgresPaymentRepository{db}
}
