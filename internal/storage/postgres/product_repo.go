package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/types"
	"gorm.io/gorm"
)


type ProductRepo struct {
	DB *Postgres
}

func NewProductRepo(postgres *Postgres) *ProductRepo {
	return &ProductRepo{
		DB: postgres,
	}
}


func (r *ProductRepo) Create(ctx context.Context, prod *types.Product) (*types.Product, error) {
	logTag := "[ProductRepo][Create]"
	log.InfofWithContext(ctx, logTag+ " creating product", "product", prod)

	db := r.DB.Cluster.GetMasterDB(ctx)

	if err := db.Create(prod).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when creating product")
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" product created successfully")
	return prod, nil
}

func (r *ProductRepo) SearchById(ctx context.Context, id int64) (*types.Product, error){
	logTag := "[ProductRepo][SearchById]"
	log.InfofWithContext(ctx, logTag+ " fetching product from db", "id", id)

	db := r.DB.Cluster.GetSlaveDB(ctx)

	var product *types.Product
	if err := db.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WarnfWithContext(ctx, logTag+" product not found", "id", id)
            return nil, fmt.Errorf("product not found with id %d", id)
		}
		log.ErrorfWithContext(ctx, logTag+" error when search by id", err.Error())
		return nil, err
	}

	return product, nil
}

func (r *ProductRepo) Search(ctx context.Context, name, category string, limit, offset int) ([]*types.Product, int64, error){
	logTag := "[ProductRepo][Create]"
	log.InfofWithContext(ctx, logTag+ " fetching product from db", "name", name, "category", category)

	db := r.DB.Cluster.GetSlaveDB(ctx)
	
	var total int64
	var products []*types.Product
	
	//build query
	query := db.Model(&types.Product{})
	if name != "" && category != "" {
		query = query.Where("name ILIKE ? AND category ILIKE ?", "%"+name+"%", "%"+category+"%")
	}else if name != ""{
		query = query.Where("name ILIKE ?", "%"+name+"%")
	}else {
		query = query.Where("category ILIKE ?", "%"+category+"%")
	}

	if err := query.Count(&total).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when fetching produc details", err.Error())
		return nil, 0, err
	}

	if err := query.Limit(limit).Offset(offset).Order("created_at DESC").Find(&products).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to search products", err)
        return nil, 0, fmt.Errorf("failed to search products: %w", err)
	}

	log.InfofWithContext(ctx, logTag+" products search completed", "found_count", len(products), "total", total)
	return products, total, nil
}

func (r *ProductRepo) GetAll(ctx context.Context, limit, offset int) ([]*types.Product, int64, error){
	logTag := "[ProductRepo][GetAll]"
    log.InfofWithContext(ctx, logTag+" fetching all products", "limit", limit, "offset", offset)

	db := r.DB.Cluster.GetSlaveDB(ctx)

	var total int64
	var products []*types.Product
	if err := db.Model(&types.Product{}).Count(&total).Order("created_at DESC").Find(&products).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to count products", err)
        return nil, 0, fmt.Errorf("failed to count products: %w", err)
	}

	if err := db.Limit(limit).Offset(offset).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to fetch products", err)
        return nil, 0, fmt.Errorf("failed to fetch products: %w", err)
	}

	return products, total, nil
}

func (r *ProductRepo) Update(ctx context.Context, prod *types.Product) (*types.Product, error){
	logTag := "[ProductRepo][Update]"
	log.InfofWithContext(ctx, logTag+" updating product", "product", prod)

	db := r.DB.Cluster.GetMasterDB(ctx)

	prod.UpdatedAt = time.Now()

	if err := db.Save(prod).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when updating the product")
		return nil, err
	}

	log.InfofWithContext(ctx, logTag+" product details updated successfully")
	return prod, nil
}

func (r *ProductRepo) UpdateStock(tx *gorm.DB, ctx context.Context, id int64, quantity int64, operation string) error{
	logTag := "[ProductRepo][UpdateStock]"
    log.InfofWithContext(ctx, logTag+" updating stock", "product_id", id, "quantity", quantity, "operation", operation)


	var product types.Product
	if err := tx.Where("id = ?", id).First(&product).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.WarnfWithContext(ctx, logTag+" product not found", "product_id", id)
            return fmt.Errorf("product not found")
        }
        log.ErrorfWithContext(ctx, logTag+" failed to fetch product", err, "product_id", id)
        return fmt.Errorf("failed to fetch product %w", err)
    }

	switch operation{
	case "set":
		product.StockQuantity = quantity
	case "add":
		product.StockQuantity += quantity
	case "subtract":
		if product.StockQuantity < quantity {
			return fmt.Errorf("insufficient stock")
		}
		product.StockQuantity -= quantity
	default:
		return fmt.Errorf("invalid operation: %s", operation)
	}

	product.UpdatedAt = time.Now()

    if err := tx.Save(&product).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to update stock", err, "product_id", id)
        return fmt.Errorf("failed to update stock %w", err)
    }

    log.InfofWithContext(ctx, logTag+" stock updated successfully", "product_id", id, "new_stock", product.StockQuantity)
    return nil

}

func (r *ProductRepo) Delete(ctx context.Context, id int64) error {
	logTag := "[ProductRepo][Delete]"
	log.InfofWithContext(ctx, logTag+" deleting product", "id", id)

	db := r.DB.Cluster.GetMasterDB(ctx)

	res := db.Where("id = ?", id).Delete(&types.Product{})
    if res.Error != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to delete product", res.Error, "product_id", id)
        return fmt.Errorf("failed to delete product %w", res.Error)
    }

	if res.RowsAffected == 0 {
		log.WarnfWithContext(ctx, logTag+" product not found", "product_id", id)
        return fmt.Errorf("product not found")
	}

	log.InfofWithContext(ctx, logTag+" product deleted successfully", "id", id)
	return nil
}

