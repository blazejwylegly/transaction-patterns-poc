package db

import (
	"github.com/blazejwylegly/transactions-poc/analytics-api/src/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

func InitDbConnection(dbConfig config.DatabaseConfig) *gorm.DB {
	//db, err := gorm.Open(postgres.Open(dbConfig.GetDbUrl()), &gorm.Config{
	//	Logger: logger.Default.LogMode(logger.Info),
	//})
	db, err := gorm.Open(postgres.Open(dbConfig.GetDbUrl()))

	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&Transaction{}, &TransactionStep{}); err != nil {
		log.Fatal(err)
	}
	return db
}
