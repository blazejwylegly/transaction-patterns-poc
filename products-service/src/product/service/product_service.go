package service

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/product/data"
	models2 "github.com/blazejwylegly/transactions-poc/products-service/src/product/models"
)

type ProductService struct {
	productRepo data.ProductRepository
}

func (service ProductService) FindAll() []*models2.ProductDto {
	var products []*models2.ProductDto
	for _, entity := range service.productRepo.FindAll() {
		productDto := &models2.ProductDto{
			ProductId:   entity.ProductID,
			Name:        entity.Name,
			Price:       entity.Price,
			Description: entity.Description,
			Quantity:    entity.Quantity,
		}
		products = append(products, productDto)
	}
	return products
}

func (service ProductService) SaveAll(products []*models2.ProductDto) {
	for _, product := range products {
		productEntity := &models2.Product{
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
			Quantity:    product.Quantity,
		}
		service.productRepo.SaveOrUpdate(productEntity)
	}
}

func (service ProductService) SaveOrUpdate(product *models2.ProductDto) {
	productEntity := &models2.Product{
		Name:        product.Name,
		Price:       product.Price,
		Description: product.Description,
		Quantity:    product.Quantity,
	}
	service.productRepo.SaveOrUpdate(productEntity)
}

func (service ProductService) FindById(productId int) *models2.ProductDto {
	entity := service.productRepo.FindById(productId)
	return &models2.ProductDto{
		ProductId:   entity.ProductID,
		Name:        entity.Name,
		Price:       entity.Price,
		Description: entity.Description,
		Quantity:    entity.Quantity,
	}
}

func (service ProductService) UpdateQuantity(productId int, body struct {
	Quantity int `json:"quantity"`
}) {
	service.productRepo.UpdateQuantity(productId, body.Quantity)
}

func NewProductService(productRepo *data.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: *productRepo,
	}
}
