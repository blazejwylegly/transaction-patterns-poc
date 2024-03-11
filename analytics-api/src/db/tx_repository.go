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
			timeI, err := time.Parse(time.RFC3339Nano, transaction.Steps[i].StepStartedAt)
			if err != nil {
				log.Printf("Error parsing date %v", err)
			}
			timeJ, err := time.Parse(time.RFC3339Nano, transaction.Steps[j].StepStartedAt)
			if err != nil {
				log.Printf("Error parsing date %v", err)
			}
			return timeI.Before(timeJ)
		})
	}

	if err != nil {
		log.Fatal(err)
	}
	return transactions
}

func (repo *TxRepository) Purge() {
	err := repo.db.Exec("DELETE FROM transaction_steps").Error
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Trunacted transaction_steps table")
	err = repo.db.Exec("DELETE FROM transactions").Error
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("Trunacted transactions table")

}
