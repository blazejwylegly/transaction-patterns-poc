package database

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func InitDbConnection(dbConfig config.DatabaseConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbConfig.GetDbUrl()))

	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&Customer{}, &Payment{}); err != nil {
		log.Fatal(err)
	}
	return db
}
