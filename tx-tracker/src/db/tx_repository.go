package db

import (
	"gorm.io/gorm"
	"log"
	"sort"
	"time"
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

	for _, transaction := range transactions {
		sort.Slice(transaction.Steps, func(i, j int) bool {
			timeI, _ := time.Parse(time.RFC3339, transaction.Steps[i].StepStartedAt)
			timeJ, _ := time.Parse(time.RFC3339, transaction.Steps[j].StepStartedAt)
			return timeI.Before(timeJ)
		})
	}

	if err != nil {
		log.Fatal(err)
	}
	return transactions
}
