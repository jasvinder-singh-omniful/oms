package setup

import (
	"github.com/omniful/go_commons/http"
	"github.com/si/internal/http/handlers"
)

func SetupRoutes(server *http.Server, userHandler *handlers.UserHandler, productHandler *handlers.ProductHandler, orderHandler *handlers.OrderHandler) {
    v1 := server.Group("/api/v1")
    {
        //user routes
        userRoutes := v1.Group("/users")
        {
			userRoutes.POST("", userHandler.CreateUserHandler)
			userRoutes.POST("/email", userHandler.GetUserByEmailHandler)
			userRoutes.POST("/id", userHandler.GetUserByIdHandler)
			userRoutes.PUT("", userHandler.UpdateUserHandler)
			userRoutes.DELETE("", userHandler.DeleteUserHandler)
        }

        //product routes
        productRoutes := v1.Group("/products")
        {
            productRoutes.POST("", productHandler.CreateProductHandler)
            productRoutes.GET("/:id", productHandler.GetProductByIdHandler)
            productRoutes.GET("", productHandler.GetAllProductsHandler)
            productRoutes.POST("/search", productHandler.SearchProductsHandler)
            productRoutes.PUT("/:id", productHandler.UpdateProductHandler)
            productRoutes.DELETE("/:id", productHandler.DeleteProductHandler)
            productRoutes.PATCH("/:id/inventory", productHandler.UpdateInventoryHandler)
        }

        //order routes
        orderRoutes := v1.Group("/orders")
        {
            orderRoutes.POST("", orderHandler.CreateOrderHandler)
            orderRoutes.GET("/:id", orderHandler.GetOrderByIdHandler)
            orderRoutes.POST("/search", orderHandler.SearchOrdersHandler)
            orderRoutes.PATCH("/:id/status", orderHandler.UpdateOrderStatusHandler)
            
            orderRoutes.POST("/:id/items", orderHandler.AddOrderItemHandler)
            orderRoutes.PUT("/:id/items/:item_id", orderHandler.UpdateOrderItemHandler)
            orderRoutes.DELETE("/:id/items/:item_id", orderHandler.RemoveOrderItemHandler)
        }
    }
}
