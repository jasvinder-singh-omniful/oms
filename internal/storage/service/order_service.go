package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/storage/postgres"
	"github.com/si/internal/types"
)

type OrderService struct {
	OrderRepo *postgres.OrderRepo
	UserRepo *postgres.UserRepo
	ProductRepo *postgres.ProductRepo
}

func NewOrderService(orderRepo *postgres.OrderRepo, userRepo *postgres.UserRepo, productRepo *postgres.ProductRepo) *OrderService{
	return &OrderService{
		OrderRepo: orderRepo,
		UserRepo: userRepo,
		ProductRepo: productRepo,
	}
}

func (s *OrderService) CreateOrder(ctx context.Context, userID int64, items []types.OrderItemRequest) (*types.OrderWithDetails, error) {
	logTag := "[OrderService][CreateOrder]"
    log.InfofWithContext(ctx, logTag+" creating order", "user_id", userID, "items_count", len(items))

	db := s.OrderRepo.DB.Cluster.GetMasterDB(ctx)
	tx := db.Begin()

	if tx.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to start transaction %w", tx.Error)
	}


	//if user exists
	_, err := s.UserRepo.SearchByID(ctx, userID)
	if err != nil {
		tx.Rollback()
		log.ErrorfWithContext(ctx, logTag+" error when getting user", err)
        return nil, err
	}

	var total float64
	var orderItems []types.OrderItem

	for _, item := range items {
		product, err := s.ProductRepo.SearchById(ctx, item.ProductID)
		if err != nil {
			tx.Rollback()
            log.ErrorfWithContext(ctx, logTag+" product not found", err, "product_id", item.ProductID)
            return nil, err
        }

		if product.StockQuantity < int64(item.Quantity){
			tx.Rollback()
			log.ErrorfWithContext(ctx, logTag+" insufficient stock", "product_id", item.ProductID, "required", item.Quantity, "available", product.StockQuantity)
            return nil, fmt.Errorf("insufficient stock")
		}

		total += float64(product.Price) * float64(item.Quantity)

		orderItems = append(orderItems, types.OrderItem{
			ProductID: product.ID,
			Quantity: item.Quantity,
			Price: product.Price,
			Name: product.Name,
		})
	}

	order := &types.Order{
		UserID: userID,
		Status: types.OrderStatusPending,
		TotalAmount: total,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	
	createdOrder, err := s.OrderRepo.CreateWithTx(tx, ctx, order, orderItems)
	if err != nil {
		tx.Rollback()
        log.ErrorfWithContext(ctx, logTag+" error when creating order", err)
        return nil, err
    }


	//update stocks
	for _, item := range items{
		err := s.ProductRepo.UpdateStock(tx, ctx, item.ProductID, int64(item.Quantity), "subtract")
		if err != nil {
			tx.Rollback()
			log.ErrorfWithContext(ctx, logTag+" error when updating stock", err, "product_id", item.ProductID)
			return nil, err
		}
	}

	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		return nil, fmt.Errorf("failed to comit transaction %v", err)
	}

	return &types.OrderWithDetails{
		Order: *createdOrder,
		Items: orderItems,
		User: nil,
	}, nil
}

func (s *OrderService) SearchOrders(ctx context.Context, params types.OrderSearchParams) ([]*types.OrderWithDetails, int64, error){
	logTag := "[OrderService][SearchOrders]"
	log.InfofWithContext(ctx, logTag+" searching orders", "params", fmt.Sprintf("%+v", params))

	orders, total, err := s.OrderRepo.SearchOrders(ctx, params)

	if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when searching orders", err)
        return nil, 0, err
    }

    log.InfofWithContext(ctx, logTag+" orders search completed", "found_count", len(orders), "total", total)
    return orders, total, nil	
}

func (s *OrderService) GetOrderById(ctx context.Context, id int64) (*types.Order, error) {
	logTag := "[OrderService][GetOrderById]"
    log.InfofWithContext(ctx, logTag+" updating order status", "order_id", id)

	order, err := s.OrderRepo.SearchByID(ctx, id)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when getting order by id", err)
        return nil, err
	}

	log.InfofWithContext(ctx, logTag+" get order by id done successfully")
	return  order, nil
}

func (s *OrderService) UpdateOrderStatus(ctx context.Context, id int64, status types.OrderStatus) (*types.Order, error) {
	logTag := "[OrderService][UpdateOrderStatus]"
    log.InfofWithContext(ctx, logTag+" updating order status", "order_id", id, "status", status)

	existingOrder, err := s.OrderRepo.SearchByID(ctx, id)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when getting existing order", err)
        return nil, err
    }

	existingOrder.Status = status
    existingOrder.UpdatedAt = time.Now()

	updatedOrder, err := s.OrderRepo.Update(ctx, existingOrder)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when updating order", err)
        return nil, fmt.Errorf("failed to update order: %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order status updated successfully", "order_id", updatedOrder.ID, "status", updatedOrder.Status)
    return updatedOrder, nil
}

func (s *OrderService) AddOrderItem(ctx context.Context, orderID, productID int64, quantity int32) (*types.OrderItem, error) {
    logTag := "[OrderService][AddOrderItem]"
    log.InfofWithContext(ctx, logTag+" adding order item", "order_id", orderID, "product_id", productID, "quantity", quantity)

	db := s.OrderRepo.DB.Cluster.GetMasterDB(ctx)

	tx := db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error when starting transaction")
	}

    _, err := s.OrderRepo.SearchByID(ctx, orderID)
    if err != nil  {
        return nil, err
    }

    product, err := s.ProductRepo.SearchById(ctx, productID)
    if err != nil  {
        return nil, err
    }

    if product.StockQuantity < int64(quantity) {
        return nil, fmt.Errorf("insufficient stock")
    }

    orderItem := &types.OrderItem{
        OrderID:   orderID,
        ProductID: productID,
        Quantity:  quantity,
        Price:     product.Price,
    }

    createdItem, err := s.OrderRepo.AddOrderItem(tx, ctx, orderItem)
    if err != nil {
		tx.Rollback()
        log.ErrorfWithContext(ctx, logTag+" error when adding order item", err)
        return nil, fmt.Errorf("failed to add order item: %w", err)
    }

    err = s.ProductRepo.UpdateStock(tx, ctx, productID, int64(quantity), "subtract")
    if err != nil {
		tx.Rollback()
        log.ErrorfWithContext(ctx, logTag+" error when updating stock", err)
    }

    err = s.OrderRepo.RecalculateOrderTotal(tx, ctx, orderID)
    if err != nil {
		tx.Rollback()
        log.ErrorfWithContext(ctx, logTag+" error when recalculating order total", err)
    }

	//commit all changes
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.ErrorfWithContext(ctx, logTag+" error when commiting changes to database")
	}

    log.InfofWithContext(ctx, logTag+" order item added successfully", "item_id", createdItem.ID)
    return createdItem, nil
}


func (s *OrderService) UpdateOrderItem(ctx context.Context, orderID, itemID int64, quantity int32) (*types.OrderItem, error) {
    logTag := "[OrderService][UpdateOrderItem]"
    log.InfofWithContext(ctx, logTag+" updating order item", "order_id", orderID, "item_id", itemID, "quantity", quantity)

	db := s.OrderRepo.DB.Cluster.GetMasterDB(ctx)

	tx := db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		return nil, fmt.Errorf("error when starting transaction")
	}

    existingItem, err := s.OrderRepo.GetOrderItem(ctx, orderID, itemID)
    if err != nil {
		tx.Rollback()
        return nil, err
    }

    product, err := s.ProductRepo.SearchById(ctx, existingItem.ProductID)
    if err != nil {
		tx.Rollback()
        return nil, err
    }

    stockDifference := int64(quantity) - int64(existingItem.Quantity)
    if stockDifference > 0 && product.StockQuantity < stockDifference {
		tx.Rollback()
        return nil, fmt.Errorf("insufficient stock")
    }

    existingItem.Quantity = quantity
    updatedItem, err := s.OrderRepo.UpdateOrderItem(ctx, existingItem)
    if err != nil {
		tx.Rollback()
        log.ErrorfWithContext(ctx, logTag+" error when updating order item", err)
        return nil, fmt.Errorf("failed to update order item: %w", err)
    }

    if stockDifference != 0 {
        operation := "subtract"
        if stockDifference < 0 {
            operation = "add"
            stockDifference = -stockDifference
        }
        err = s.ProductRepo.UpdateStock(tx, ctx, existingItem.ProductID, stockDifference, operation)
        if err != nil {
			tx.Rollback()
            log.ErrorfWithContext(ctx, logTag+" error when updating stock", err)
        }
    }

    err = s.OrderRepo.RecalculateOrderTotal(tx, ctx, orderID)
    if err != nil {
		tx.Rollback()
        log.ErrorfWithContext(ctx, logTag+" error when recalculating order total", err)
    }

	//commit all changes
	if err := tx.Commit().Error; err != nil {
		tx.Rollback()
		log.ErrorfWithContext(ctx, logTag+" error when commiting changes to database")
	}

    log.InfofWithContext(ctx, logTag+" order item updated successfully", "item_id", updatedItem.ID)
    return updatedItem, nil
}

func (s *OrderService) RemoveOrderItem(ctx context.Context, orderID, itemID int64) error {
    logTag := "[OrderService][RemoveOrderItem]"
    log.InfofWithContext(ctx, logTag+" removing order item", "order_id", orderID, "item_id", itemID)

	db := s.OrderRepo.DB.Cluster.GetMasterDB(ctx)

	tx := db.Begin()
	if tx.Error != nil {
		tx.Rollback()
		return errors.New("error when starting transaction")
	}

    existingItem, err := s.OrderRepo.GetOrderItem(ctx, orderID, itemID)
    if err != nil || existingItem == nil {
        return errors.New("order item not found")
    }

    err = s.OrderRepo.RemoveOrderItem(tx,ctx, orderID, itemID)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when removing order item", err)
        return fmt.Errorf("failed to remove order item: %w", err)
    }

    err = s.ProductRepo.UpdateStock(tx, ctx, existingItem.ProductID, int64(existingItem.Quantity), "add")
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when restoring stock", err)
    }

    err = s.OrderRepo.RecalculateOrderTotal(tx, ctx, orderID)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when recalculating order total", err)
    }

    log.InfofWithContext(ctx, logTag+" order item removed successfully", "item_id", itemID)
    return nil
}


