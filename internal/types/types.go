package types

import (
	"time"

	"github.com/omniful/go_commons/http"
)

type ErrorResponse struct {
	Message string        `json:"message"`
	Details []interface{} `json:"details,omitempty"`
}

type APIResponse struct {
	Headers http.ResponseParams `json:"headers,omitempty"`
	Message string              `json:"message"`
	Data    interface{}         `json:"data,omitempty"`
	Error   *ErrorResponse      `json:"error,omitempty"`
}

type User struct {
	ID int64 `json:"id" db:"id"`

	Name         string `json:"name" db:"name"`
	Email        string `json:"email" db:"email"`
	Phone        string `json:"phone" db:"phone"`
	PasswordHash string `json:"-" db:"password_hash"`
	IsActive     bool   `json:"is_active" db:"is_active"`

	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type Product struct {
	ID int64 `json:"id" db:"id"`

	Name          string  `json:"name" db:"name"`
	SKU           string  `json:"sku" db:"sku"`
	Price         float64 `json:"price" db:"price"`
	Category      string  `json:"category" db:"category"`
	StockQuantity int64   `json:"stock_quantity" db:"stock_quantity"`

	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
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
	ID          int64       `json:"id" db:"id"`
	UserID      int64       `json:"user_id" db:"user_id"`
	Status      OrderStatus `json:"status" db:"status"`
	TotalAmount float64     `json:"total_amount" db:"total_amount"`

	CreatedAt time.Time `json:"created_at,omitempty" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at,omitempty" db:"updated_at"`
}

type OrderItem struct {
	ID        int64 `json:"id" db:"id"`
	OrderID   int64 `json:"order_id" db:"order_id"`
	ProductID int64 `json:"product_id" db:"product_id"`

	Name     string  `json:"name" db:"name"`
	Quantity int32   `json:"quantity" db:"quantity"`
	Price    float64 `json:"price" db:"price"`
}
