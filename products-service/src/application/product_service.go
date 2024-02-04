package application

import (
	"github.com/blazejwylegly/transactions-poc/products-service/src/database"
	"github.com/blazejwylegly/transactions-poc/products-service/src/dto"
	"github.com/google/uuid"
	"log"
)

type ProductService struct {
	productRepo database.ProductRepository
}

func (service *ProductService) FindAll() []*dto.ProductDto {
	var products []*dto.ProductDto
	for _, entity := range service.productRepo.FindAll() {
		productDto := &dto.ProductDto{
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

func (service *ProductService) SaveAll(products []*dto.ProductDto) {
	for _, product := range products {
		productEntity := &database.Product{
			Name:        product.Name,
			Price:       product.Price,
			Description: product.Description,
			Quantity:    product.Quantity,
		}
		service.productRepo.SaveOrUpdate(productEntity)
	}
}

func (service *ProductService) SaveOrUpdate(product *dto.ProductDto) {
	productId, err := uuid.NewUUID()
	if err != nil {
		log.Printf("Error creating uuid for new product: %v!", err)
	}
	productEntity := &database.Product{
		ProductID:   productId,
		Name:        product.Name,
		Price:       product.Price,
		Description: product.Description,
		Quantity:    product.Quantity,
	}
	service.productRepo.SaveOrUpdate(productEntity)
}

func (service *ProductService) FindById(productId uuid.UUID) *dto.ProductDto {
	entity := service.productRepo.FindById(productId)
	return &dto.ProductDto{
		ProductId:   entity.ProductID,
		Name:        entity.Name,
		Price:       entity.Price,
		Description: entity.Description,
		Quantity:    entity.Quantity,
	}
}

func (service *ProductService) UpdateQuantity(productId uuid.UUID, body struct {
	Quantity int `json:"quantity"`
}) {
	service.productRepo.UpdateQuantity(productId, body.Quantity)
}

func NewProductService(productRepo database.ProductRepository) *ProductService {
	return &ProductService{
		productRepo: productRepo,
	}
}
