package data

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/models"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type ProductRepository interface {
	FindAll() []*models.Product
	SaveOrUpdate(*models.Product)
	FindById(uuid.UUID) *models.Product
	UpdateQuantity(uuid.UUID, int)
}

type postgresProductRepository struct {
	db *gorm.DB
}

func (postgresRepo postgresProductRepository) UpdateQuantity(productId uuid.UUID, quantity int) {
	var product *models.Product
	postgresRepo.db.Model(&product).Where("product_id = ?", productId).Update("quantity", quantity)
}

func (postgresRepo postgresProductRepository) FindAll() []*models.Product {
	var products []*models.Product
	postgresRepo.db.Find(&products)
	return products
}

func (postgresRepo postgresProductRepository) FindById(productId uuid.UUID) *models.Product {
	var product *models.Product
	postgresRepo.db.First(&product, productId)
	return product
}

func (postgresRepo postgresProductRepository) SaveOrUpdate(product *models.Product) {
	postgresRepo.db.Save(product)
}

func NewProductRepository(db *gorm.DB) ProductRepository {
	return &postgresProductRepository{db}
}
