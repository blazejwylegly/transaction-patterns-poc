package database

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll() []*Product
	SaveOrUpdate(*Product)
	FindById(uuid.UUID) *Product
	UpdateQuantity(uuid.UUID, int)
}

type PostgresProductRepository struct {
	db *gorm.DB
}

func (repo *PostgresProductRepository) UpdateQuantity(productId uuid.UUID, quantity int) {
	var product *Product
	repo.db.Model(&product).Where("product_id = ?", productId).Update("quantity", quantity)
}

func (repo *PostgresProductRepository) FindAll() []*Product {
	var products []*Product
	repo.db.Find(&products)
	return products
}

func (repo *PostgresProductRepository) FindById(productId uuid.UUID) *Product {
	var product *Product
	repo.db.First(&product, productId)
	return product
}

func (repo *PostgresProductRepository) SaveOrUpdate(product *Product) {
	repo.db.Save(product)
}

func NewProductRepository(db *gorm.DB) *PostgresProductRepository {
	return &PostgresProductRepository{db}
}
