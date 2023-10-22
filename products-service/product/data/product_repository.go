package data

import (
	"fmt"
	"github.com/blazejwylegly/transactions-poc/products-service/config"
	"github.com/blazejwylegly/transactions-poc/products-service/product/models"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"log"
)

type ProductRepository interface {
	FindAll() []*models.Product
	SaveOrUpdate(*models.Product)
	FindById(int) *models.Product
	UpdateQuantity(int, int)
}

type postgresProductRepository struct {
	db *gorm.DB
}

func (postgresRepo postgresProductRepository) UpdateQuantity(productId int, quantity int) {
	var product *models.Product
	postgresRepo.db.Model(&product).Where("product_id = ?", productId).Update("quantity", quantity)
}

func (postgresRepo postgresProductRepository) FindAll() []*models.Product {
	var products []*models.Product
	postgresRepo.db.Find(&products)
	return products
}

func (postgresRepo postgresProductRepository) FindById(productId int) *models.Product {
	var product *models.Product
	postgresRepo.db.First(&product, productId)
	return product
}

func (postgresRepo postgresProductRepository) SaveOrUpdate(product *models.Product) {
	postgresRepo.db.Save(product)
}

func NewProductRepository(dbConfig config.DatabaseConfig) ProductRepository {
	return &postgresProductRepository{
		db: initDbConnection(dbConfig),
	}
}

func initDbConnection(dbConfig config.DatabaseConfig) *gorm.DB {
	dbUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s",
		dbConfig.DbUsername,
		dbConfig.DbPassword,
		dbConfig.DbHost,
		dbConfig.DbPort,
		dbConfig.DbName,
	)
	db, err := gorm.Open(postgres.Open(dbUrl), &gorm.Config{})
	if err != nil {
		log.Fatal(err)
	}

	if err := db.AutoMigrate(&models.Product{}); err != nil {
		log.Fatal(err)
	}
	return db
}
