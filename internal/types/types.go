package types

import (
	"time"
)

type ErrorResponse struct {
	Details []interface{} `json:"details,omitempty"`
}

type APIResponse struct {
	Message string         `json:"message"`
	Data    interface{}    `json:"data,omitempty"`
	Error   *ErrorResponse `json:"error,omitempty"`
}

type User struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement"`

	Name         string `json:"name" gorm:"column:name;not null"`
	Email        string `json:"email" gorm:"column:email;unique;not null"`
	Phone        string `json:"phone" gorm:"column:phone;unique"`
	PasswordHash string `json:"-" gorm:"column:password_hash;not null"`
	IsActive     bool   `json:"is_active" gorm:"column:is_active;default:true"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"column:updated_at;autoUpdateTime"`
}

type Product struct {
	ID int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement"`

	Name          string  `json:"name" gorm:"column:name;not null"`
	SKU           string  `json:"sku" gorm:"column:sku;unique;not null"`
	Price         float64 `json:"price" gorm:"column:price;not null"`
	Category      string  `json:"category" gorm:"column:category"`
	StockQuantity int64   `json:"stock_quantity" gorm:"column:stock_quantity;default:0"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"column:updated_at;autoUpdateTime"`
}

// enum type OrderStatus
type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "order.pending"
	OrderStatusShipped   OrderStatus = "order.shipped"
	OrderStatusCancelled OrderStatus = "order.cancelled"
	OrderStatusDelivered OrderStatus = "order.delivered"
)

type Order struct {
	ID          int64       `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	UserID      int64       `json:"user_id" gorm:"column:user_id;not null;index"`
	Status      OrderStatus `json:"status" gorm:"column:status;type:order_status;default:'order.pending'"`
	TotalAmount float64     `json:"total_amount" gorm:"column:total_amount;not null;default:0"`

	CreatedAt time.Time `json:"created_at,omitempty" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt time.Time `json:"updated_at,omitempty" gorm:"column:updated_at;autoUpdateTime"`
}

type OrderItem struct {
	ID        int64 `json:"id" gorm:"column:id;primaryKey;autoIncrement"`
	OrderID   int64 `json:"order_id" gorm:"column:order_id;not null;index"`
	ProductID int64 `json:"product_id" gorm:"column:product_id;not null;index"`

	Name     string  `json:"name" gorm:"column:name;not null"`
	Quantity int32   `json:"quantity" gorm:"column:quantity;not null"`
	Price    float64 `json:"price" gorm:"column:price;not null"`
}
