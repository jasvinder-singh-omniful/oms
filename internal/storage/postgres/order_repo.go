package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/omniful/go_commons/log"
	"github.com/si/internal/types"
	"gorm.io/gorm"
)

type OrderRepo struct {
	DB *Postgres
}

func NewOrderRepo(db *Postgres) *OrderRepo {
	return &OrderRepo{
		DB: db,
	}
}

func (r *OrderRepo) CreateWithTx(tx *gorm.DB,  ctx context.Context, order *types.Order, orderItems []types.OrderItem) (*types.Order, error){
	logTag := "[OrderRepo][Create]"
    log.InfofWithContext(ctx, logTag+" creating order", "user_id", order.UserID, "items_count", len(orderItems))
	

	if tx.Error != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to begin transaction", tx.Error)
		return nil, fmt.Errorf("failed to begin transaction %w", tx.Error)
	}

	//create order in db
	if err := tx.Create(order).Error; err != nil {
		tx.Rollback()
		log.ErrorfWithContext(ctx, logTag+" failed to create order", err, "user_id", order.UserID)
		return nil, fmt.Errorf("failed to create order %v", err)
	}

	//create order items
	for i := range orderItems {
		orderItems[i].OrderID = order.ID
		if err := tx.Create(&orderItems[i]).Error; err != nil {
			tx.Rollback()
			log.ErrorfWithContext(ctx, logTag+" failed to create order item", err, "product_id", orderItems[i].ProductID)
            return nil, fmt.Errorf("failed to create order item %v", err)
		}
	}

  

	log.InfofWithContext(ctx, logTag+" order created successfully", "order_id", order.ID)
    return order, nil
}

func (r *OrderRepo) SearchByID(ctx context.Context, id int64) (*types.Order, error) {
    logTag := "[OrderRepo][SearchByID]"
    log.InfofWithContext(ctx, logTag+" fetching order", "order_id", id)

    db := r.DB.Cluster.GetSlaveDB(ctx)

    var order types.Order
    if err := db.Where("id = ?", id).First(&order).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            log.WarnfWithContext(ctx, logTag+" order not found", "order_id", id)
            return nil, fmt.Errorf("order not found")
        }
        log.ErrorfWithContext(ctx, logTag+" failed to fetch order", err, "order_id", id)
        return nil, fmt.Errorf("failed to fetch order %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order fetched successfully", "order_id", order.ID)
    return &order, nil
}

func (r *OrderRepo) GetOrderDetails(ctx context.Context, orderId int64) (*types.OrderWithDetails, error) {
	logTag := "[OrderRepo][GetOrderWithDetails]"
    log.InfofWithContext(ctx, logTag+" fetching order with details", "order_id", orderId)

	db := r.DB.Cluster.GetSlaveDB(ctx)

	//get order
	var order types.Order
	if err := db.Where("id = ?", orderId).First(&order).Error; err != nil {
		if err == gorm.ErrRecordNotFound{
			log.WarnfWithContext(ctx, logTag+" order not found", "order_id", orderId)
            return nil, fmt.Errorf("order not found")
        }
        log.ErrorfWithContext(ctx, logTag+" failed to fetch order", err, "order_id", orderId)
        return nil, fmt.Errorf("failed to fetch order %w", err)
	}

	//get order items
	var orderItems []types.OrderItem
	if err := db.Where("order_id = ?", orderId).Find(&orderItems).Error; err != nil {
		log.ErrorfWithContext(ctx, logTag+" failed to fetch order items", err, "order_id", orderId)
        return nil, fmt.Errorf("failed to fetch order items: %w", err)
	}

	//get user details form  user_id
	var user types.User
    if err := db.Where("id = ?", order.UserID).First(&user).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to fetch user", err, "user_id", order.UserID)
        return nil, fmt.Errorf("failed to fetch user: %w", err)
    }

	orderWithDetails := &types.OrderWithDetails{
        Order: order,
        Items: orderItems,
        User:  &user,
    }

    log.InfofWithContext(ctx, logTag+" order with details fetched successfully", "order_id", order.ID, "items_count", len(orderItems))
    return orderWithDetails, nil
}


func (r *OrderRepo) SearchOrders(ctx context.Context, params types.OrderSearchParams) ([]*types.OrderWithDetails, int64, error) {
    logTag := "[OrderRepo][SearchOrders]"
    log.InfofWithContext(ctx, logTag+" searching orders", "params", fmt.Sprintf("%+v", params))

    db := r.DB.Cluster.GetSlaveDB(ctx)

    baseQuery := db.Table("orders o").
        Joins("JOIN users u ON o.user_id = u.id")

    //filters
    if params.UserID != 0 {
        baseQuery = baseQuery.Where("o.user_id = ?", params.UserID)
    }
    if params.OrderID != 0 {
        baseQuery = baseQuery.Where("o.id = ?", params.OrderID)
    }
    if params.CustomerName != "" {
        baseQuery = baseQuery.Where("u.Name ILIKE ?", params.CustomerName)
    }
    if params.Status != "" {
        baseQuery = baseQuery.Where("o.status = ?", params.Status)
    }

    //filter by item name
    if params.ItemName != "" {
        subQuery := db.Table("order_items oi").
            Select("oi.order_id").
            Where("oi.name ILIKE ?", params.ItemName)

        baseQuery = baseQuery.Where("o.id IN (?)", subQuery)
    }

    //count query
    var total int64
    countQuery := baseQuery.Session(&gorm.Session{}) // clean session
    if err := countQuery.Count(&total).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to count orders", err)
        return nil, 0, fmt.Errorf("failed to count orders: %w", err)
    }

    var orderResults []struct {
        types.Order
        UserName  string `json:"user_name"`
        UserEmail string `json:"user_email"`
    }

    query := baseQuery.Select("o.*, u.Name as user_name, u.Email as user_email")
    res := query.Limit(params.Limit).Offset(params.Offset).Order("o.created_at DESC").Find(&orderResults)
    
    if err := res.Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to search orders", err)
        return nil, 0, fmt.Errorf("failed to search orders: %w", err)
    }

    var orderWithDetails []*types.OrderWithDetails
    for _, orderResult := range orderResults {
        var items []types.OrderItem
        res := db.Where("order_id = ?", orderResult.ID).Find(&items)
        if err := res.Error; err != nil {
            log.ErrorfWithContext(ctx, logTag+" failed to fetch order items", err, "order_id", orderResult.ID)
            continue
        }

        user := types.User{
            ID:    orderResult.UserID,
            Name:  orderResult.UserName,
            Email: orderResult.UserEmail,
        }

        orderWithDetail := &types.OrderWithDetails{
            Order: orderResult.Order,
            Items: items,
            User:  &user,
        }
        orderWithDetails = append(orderWithDetails, orderWithDetail)
    }

    log.InfofWithContext(ctx, logTag+" orders searched successfully", "count", len(orderWithDetails), "total", total)
    return orderWithDetails, total, nil
}


func (r *OrderRepo) Update(ctx context.Context, order *types.Order) (*types.Order, error) {
    logTag := "[OrderRepo][Update]"
    log.InfofWithContext(ctx, logTag+" updating order", "order_id", order.ID)

    db := r.DB.Cluster.GetMasterDB(ctx)

    order.UpdatedAt = time.Now()

    if err := db.Save(order).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to update order", err, "order_id", order.ID)
        return nil, fmt.Errorf("failed to update order %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order updated successfully", "order_id", order.ID)
    return order, nil
}

func (r *OrderRepo) AddOrderItem(tx *gorm.DB, ctx context.Context, item *types.OrderItem) (*types.OrderItem, error) {
    logTag := "[OrderRepo][AddOrderItem]"
    log.InfofWithContext(ctx, logTag+" adding order item", "order_id", item.OrderID, "product_id", item.ProductID)


    if err := tx.Create(item).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to add order item", err, "order_id", item.OrderID)
        return nil, fmt.Errorf("failed to add order item %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order item added successfully", "item_id", item.ID)
    return item, nil
}

func (r *OrderRepo) GetOrderItem(ctx context.Context, orderID, itemID int64) (*types.OrderItem, error) {
    logTag := "[OrderRepo][GetOrderItem]"
    log.InfofWithContext(ctx, logTag+" fetching order item", "order_id", orderID, "item_id", itemID)

    db := r.DB.Cluster.GetSlaveDB(ctx)

    var orderItem types.OrderItem
    if err := db.Where("id = ? AND order_id = ?", itemID, orderID).First(&orderItem).Error; err != nil {
        if err == gorm.ErrRecordNotFound {
            log.WarnfWithContext(ctx, logTag+" order item not found", "order_id", orderID, "item_id", itemID)
            return nil, fmt.Errorf("order item not found")
        }
        log.ErrorfWithContext(ctx, logTag+" failed to fetch order item", err, "order_id", orderID, "item_id", itemID)
        return nil, fmt.Errorf("failed to fetch order item %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order item fetched successfully", "item_id", orderItem.ID)
    return &orderItem, nil
}

func (r *OrderRepo) UpdateOrderItem(ctx context.Context, item *types.OrderItem) (*types.OrderItem, error) {
    logTag := "[OrderRepo][UpdateOrderItem]"
    log.InfofWithContext(ctx, logTag+" updating order item", "item_id", item.ID)

    db := r.DB.Cluster.GetMasterDB(ctx)

    if err := db.Save(item).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to update order item", err, "item_id", item.ID)
        return nil, fmt.Errorf("failed to update order item %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order item updated successfully", "item_id", item.ID)
    return item, nil
}

func (r *OrderRepo) RemoveOrderItem(tx *gorm.DB, ctx context.Context, orderID, itemID int64) error {
    logTag := "[OrderRepo][RemoveOrderItem]"
    log.InfofWithContext(ctx, logTag+" removing order item", "order_id", orderID, "item_id", itemID)

    result := tx.Where("id = ? AND order_id = ?", itemID, orderID).Delete(&types.OrderItem{})
    if result.Error != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to remove order item", result.Error, "order_id", orderID, "item_id", itemID)
        return fmt.Errorf("failed to remove order item %w", result.Error)
    }

    if result.RowsAffected == 0 {
        log.WarnfWithContext(ctx, logTag+" order item not found", "order_id", orderID, "item_id", itemID)
        return fmt.Errorf("order item not found")
    }

    log.InfofWithContext(ctx, logTag+" order item removed successfully", "order_id", orderID, "item_id", itemID)
    return nil
}

func (r *OrderRepo) RecalculateOrderTotal(tx *gorm.DB, ctx context.Context, orderID int64) error {
    logTag := "[OrderRepo][RecalculateOrderTotal]"
    log.InfofWithContext(ctx, logTag+" recalculating order total", "order_id", orderID)


    var items []types.OrderItem
    if err := tx.Where("order_id = ?", orderID).Find(&items).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to fetch order items", err, "order_id", orderID)
        return fmt.Errorf("failed to fetch order items: %w", err)
    }

    var newTotal float64
    for _, item := range items {
        newTotal += item.Price * float64(item.Quantity)
    }

    if err := tx.Model(&types.Order{}).Where("id = ?", orderID).Update("total_amount", newTotal).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to update order total", err, "order_id", orderID)
        return fmt.Errorf("failed to update order total %w", err)
    }

    log.InfofWithContext(ctx, logTag+" order total recalculated successfully", "order_id", orderID, "new_total", newTotal)
    return nil
}

func (r *OrderRepo) GetOrdersByUserID(ctx context.Context, userID int64, limit, offset int) ([]*types.OrderWithDetails, int64, error) {
    logTag := "[OrderRepo][GetOrdersByUserID]"
    log.InfofWithContext(ctx, logTag+" fetching orders by user ID", "user_id", userID)

    db := r.DB.Cluster.GetSlaveDB(ctx)

    var total int64
    var orders []types.Order

    if err := db.Model(&types.Order{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to count orders", err)
        return nil, 0, fmt.Errorf("failed to count orders: %w", err)
    }

    if err := db.Where("user_id = ?", userID).Limit(limit).Offset(offset).Order("created_at DESC").Find(&orders).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to fetch orders", err)
        return nil, 0, fmt.Errorf("failed to fetch orders: %w", err)
    }

    var user types.User
    if err := db.Where("id = ?", userID).First(&user).Error; err != nil {
        log.ErrorfWithContext(ctx, logTag+" failed to fetch user", err, "user_id", userID)
        return nil, 0, fmt.Errorf("failed to fetch user: %w", err)
    }

    var ordersWithDetails []*types.OrderWithDetails
    for _, order := range orders {
        var items []types.OrderItem
        if err := db.Where("order_id = ?", order.ID).Find(&items).Error; err != nil {
            log.ErrorfWithContext(ctx, logTag+" failed to fetch order items", err, "order_id", order.ID)
            continue
        }

        orderWithDetails := &types.OrderWithDetails{
            Order: order,
            Items: items,
            User:  &user,
        }

        ordersWithDetails = append(ordersWithDetails, orderWithDetails)
    }

    log.InfofWithContext(ctx, logTag+" orders fetched successfully", "user_id", userID, "count", len(ordersWithDetails), "total", total)
    return ordersWithDetails, total, nil
}
