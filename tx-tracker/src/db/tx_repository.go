package db

import (
	"gorm.io/gorm"
	"log"
)

type TxRepository struct {
	db *gorm.DB
}

func NewDbRepository(dbConnection *gorm.DB) *TxRepository {
	return &TxRepository{db: dbConnection}
}

func (repo *TxRepository) GetAll() []Transaction {
	var transactions []Transaction
	err := repo.db.Model(&Transaction{}).Preload("Steps").Find(&transactions).Error

	if err != nil {
		log.Fatal(err)
	}
	return transactions
}
