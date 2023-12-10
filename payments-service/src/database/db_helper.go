package database

import (
	"github.com/blazejwylegly/transactions-poc/payments-service/src/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
)

func InitDbConnection(dbConfig config.DatabaseConfig) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dbConfig.GetDbUrl()), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&Customer{}, &Payment{}); err != nil {
		log.Fatal(err)
	}
	return db
}
