package service

import (
	"context"
	"fmt"
	"time"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/storage/postgres"
	"github.com/si/internal/types"
)

type ProductService struct {
	ProductRepo *postgres.ProductRepo
}

func NewProductService(productRepo *postgres.ProductRepo) *ProductService {
	return &ProductService{
		ProductRepo: productRepo,
	}
}

func (s *ProductService) CreateProduct(ctx context.Context, name string, sku string, price float64, category string, stockQuantity int64) (*types.Product, error) {
	logTag := "[ProductService][CreateProduct]"
	log.InfofWithContext(ctx, logTag+" creating product", "product", name)

	product := &types.Product{
		Name:          name,
		SKU:           sku,
		Price:         price,
		Category:      category,
		StockQuantity: stockQuantity,
	}

	prod, err := s.ProductRepo.Create(ctx, product)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when creating product", err)
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" product created successfully ", prod)
	return prod, nil
}


func (s *ProductService) GetProductById(ctx context.Context, id int64) (*types.Product, error){
	logTag := "[ProductService][GetByID]"
	log.InfofWithContext(ctx, logTag+" getting all products")

	product, err := s.ProductRepo.SearchById(ctx, id)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when getting products by id", err)
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" product fetched successfully", "product", product)
	return product, nil
}

func (s *ProductService) GetAllProducts(ctx context.Context, limit, offset int) ([]*types.Product, int64, error) {
	logTag := "[ProductService][GetAllProducts]"
	log.InfofWithContext(ctx, logTag+" getting all products", "limit", limit, "offset", offset)

	products, total, err := s.ProductRepo.GetAll(ctx, limit, offset)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when getting all products", err)
		return nil, 0, err
	}

	log.InfofWithContext(ctx, logTag+" products fetched successfully", "count", len(products), "total", total)
	return products, total, nil
}

func (s *ProductService) SearchProduct(ctx context.Context, name, category string, limit, offset int) ([]*types.Product, int64, error) {
	logTag := "[ProductService][SearchProducts]"
	log.InfofWithContext(ctx, logTag+" searching products", "name", name, "category", category, "limit", limit, "offset", offset)

	products, total, err := s.ProductRepo.Search(ctx, name, category, limit, offset)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when searching products", err)
		return nil, 0, err
	}

	log.InfofWithContext(ctx, logTag+"  products search completed", "found_count", len(products), "total", total)
	return products, total, nil
}

func (s *ProductService) UpdateProduct(ctx context.Context, id int64, name string, price float64, category string, stockQuantity int64) (*types.Product, error) {
	logTag := "[ProductService][UpdateProduct]"
	log.InfofWithContext(ctx, logTag+" updating product", "product_id", id)

	existingProduct, err := s.ProductRepo.SearchById(ctx, id)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when getting existing product", err)
		return nil, err
	}

	if name != "" {
		existingProduct.Name = name
	}
	if price > 0 {
		existingProduct.Price = price
	}
	if category != "" {
		existingProduct.Category = category
	}
	if stockQuantity >= 0 {
		existingProduct.StockQuantity = stockQuantity
	}
	existingProduct.UpdatedAt = time.Now()

	updatedProduct, err := s.ProductRepo.Update(ctx, existingProduct)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when updating product", err)
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" product updated successfully", "product_id", updatedProduct.ID)
	return updatedProduct, nil
}

func (s *ProductService) DeleteProduct(ctx context.Context, id int64) error {
	logTag := "[ProductService][DeleteProduct]"
    log.InfofWithContext(ctx, logTag+" deleting product", "product_id", id)

    err := s.ProductRepo.Delete(ctx, id)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when deleting product", err)
        return err
    }

    log.InfofWithContext(ctx, logTag+" product deleted successfully", "product_id", id)
    return nil
}

//updates quanity only
func (s *ProductService) UpdateInventory(ctx context.Context, id int64, quantity int64, operation string) (*types.Product, error){
	logTag := "[ProductService][UpdateInventory]"
    log.InfofWithContext(ctx, logTag+" updating inventory", "product_id", id, "quantity", quantity, "operation", operation)

	existingProduct, err := s.ProductRepo.SearchById(ctx, id)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when getting existing product", err)
        return nil, err
    }

	switch operation {
	case "set":
		existingProduct.StockQuantity = quantity
	case "add":
		existingProduct.StockQuantity += quantity
	case "subtract":
		if existingProduct.StockQuantity < quantity{
			return nil, fmt.Errorf("insufficient stock")
		}
		existingProduct.StockQuantity -= quantity
	default:
		return nil, fmt.Errorf("invalid operation %s", operation)
	}

	existingProduct.UpdatedAt = time.Now()

	updatedProduct, err := s.ProductRepo.Update(ctx, existingProduct)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when updating inventory", err)
        return nil, fmt.Errorf("failed to update inventory %w", err)
	}

	log.InfofWithContext(ctx, logTag+" inventory updated successfully", "product_id", updatedProduct.ID, "new_stock", updatedProduct.StockQuantity)
    return updatedProduct, nil
}
