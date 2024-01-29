package database

import (
	"gorm.io/gorm"
	"log"
)

type SagaRepository struct {
	db *gorm.DB
}

func NewSagaRepository(dbConnection *gorm.DB) *SagaRepository {
	return &SagaRepository{db: dbConnection}
}

func (repo *SagaRepository) GetAll() []Transaction {
	var transactions []Transaction
	err := repo.db.Model(&Transaction{}).Preload("Steps").Find(&transactions).Error

	if err != nil {
		log.Fatal(err)
	}
	return transactions
}
