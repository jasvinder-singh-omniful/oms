package handlers

import (
    "strconv"

    "github.com/gin-gonic/gin"
    "github.com/omniful/go_commons/http"
    "github.com/omniful/go_commons/log"
    "github.com/omniful/go_commons/validator"
    "github.com/si/internal/storage/service"
    "github.com/si/internal/types"
    "github.com/si/internal/utils/response"
)

type OrderHandler struct {
    OrderService *service.OrderService
}

func NewOrderHandler(orderService *service.OrderService) *OrderHandler {
    return &OrderHandler{
        OrderService: orderService,
    }
}

func (h *OrderHandler) CreateOrderHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][CreateOrderHandler]"
    log.InfofWithContext(ctx, logTag+" creating order")

    var body struct {
        UserID int64 `json:"user_id" validate:"required,numeric"`
        Items  []struct {
            ProductID int64 `json:"product_id" validate:"required,numeric"`
            Quantity  int32 `json:"quantity" validate:"required,numeric,min=1"`
        } `json:"items" validate:"required,min=1"`
    }

    if err := c.ShouldBindJSON(&body); err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when deserializing body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid request body", err.Error()))
        return
    }

    if err := validator.ValidateStruct(ctx, body); err.Exists() {
        log.ErrorfWithContext(ctx, logTag+" error when validating the body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter valid inputs", err.ErrorMap()))
        return
    }

    var orderItems []types.OrderItemRequest
    for _, item := range body.Items {
        orderItems = append(orderItems, types.OrderItemRequest{
            ProductID: item.ProductID,
            Quantity:  item.Quantity,
        })
    }

    order, err := h.OrderService.CreateOrder(ctx, body.UserID, orderItems)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when creating order", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when creating order", err.Error()))
        return
    }

    c.JSON(http.StatusCreated.Code(), gin.H{
        "message": "order created successfully",
        "order":   order,
    })
}

func (h *OrderHandler) GetOrderByIdHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][GetOrderByIdHandler]"

    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("order ID is required", ""))
        return
    }

    orderID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid order ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid order ID format", err.Error()))
        return
    }

    order, err := h.OrderService.GetOrderById(ctx, orderID)
    if err != nil {
        if err.Error() == "order not found" {
            c.JSON(http.StatusNotFound.Code(), gin.H{
                "message": "order not found",
                "order":   nil,
            })
            return
        }
        log.ErrorfWithContext(ctx, logTag+" error when getting order", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when getting order", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "order fetched successfully",
        "order":   order,
    })
}

func (h *OrderHandler) SearchOrdersHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][SearchOrdersHandler]"

    var body struct {
        UserID       int64               `json:"user_id" validate:"omitempty,numeric"`
        OrderID      int64               `json:"order_id" validate:"omitempty,numeric"`
        CustomerName string              `json:"customer_name" validate:"omitempty"`
        ItemName     string              `json:"item_name" validate:"omitempty"`
        Status       types.OrderStatus   `json:"status" validate:"omitempty,oneof=order.pending order.shipped order.cancelled order.delivered"`
        Page         int                 `json:"page" validate:"omitempty,numeric"`
        Limit        int                 `json:"limit" validate:"omitempty,numeric"`
    }

    if err := c.ShouldBindJSON(&body); err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when deserializing body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid request body", err.Error()))
        return
    }

    if err := validator.ValidateStruct(ctx, body); err.Exists() {
        log.ErrorfWithContext(ctx, logTag+" error when validating the body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter valid inputs", err.ErrorMap()))
        return
    }

    // Set defaults
    if body.Page < 1 {
        body.Page = 1
    }
    if body.Limit < 1 || body.Limit > 100 {
        body.Limit = 10
    }

    offset := (body.Page - 1) * body.Limit

    searchParams := types.OrderSearchParams{
        UserID:       body.UserID,
        OrderID:      body.OrderID,
        CustomerName: body.CustomerName,
        ItemName:     body.ItemName,
        Status:       body.Status,
        Limit:        body.Limit,
        Offset:       offset,
    }

    orders, total, err := h.OrderService.SearchOrders(ctx, searchParams)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when searching orders", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when searching orders", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "orders search completed",
        "orders":  orders,
        "total":   total,
        "page":    body.Page,
        "limit":   body.Limit,
    })
}

func (h *OrderHandler) UpdateOrderStatusHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][UpdateOrderStatusHandler]"

    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("order ID is required", ""))
        return
    }

    orderID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid order ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid order ID format", err.Error()))
        return
    }

    var body struct {
        Status types.OrderStatus `json:"status" validate:"required,oneof=order.pending order.shipped order.cancelled order.delivered"`
    }

    if err := c.ShouldBindJSON(&body); err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when deserializing body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid request body", err.Error()))
        return
    }

    if err := validator.ValidateStruct(ctx, body); err.Exists() {
        log.ErrorfWithContext(ctx, logTag+" error when validating the body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter valid inputs", err.ErrorMap()))
        return
    }

    updatedOrder, err := h.OrderService.UpdateOrderStatus(ctx, orderID, body.Status)
    if err != nil {
        if err.Error() == "order not found" {
            c.JSON(http.StatusNotFound.Code(), response.ErrorResponse("order not found", err.Error()))
            return
        }
        log.ErrorfWithContext(ctx, logTag+" error when updating order status", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when updating order status", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "order status updated successfully",
        "order":   updatedOrder,
    })
}

func (h *OrderHandler) AddOrderItemHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][AddOrderItemHandler]"

    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("order ID is required", ""))
        return
    }

    orderID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid order ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid order ID format", err.Error()))
        return
    }

    var body struct {
        ProductID int64 `json:"product_id" validate:"required,numeric"`
        Quantity  int32 `json:"quantity" validate:"required,numeric,min=1"`
    }

    if err := c.ShouldBindJSON(&body); err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when deserializing body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid request body", err.Error()))
        return
    }

    if err := validator.ValidateStruct(ctx, body); err.Exists() {
        log.ErrorfWithContext(ctx, logTag+" error when validating the body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter valid inputs", err.ErrorMap()))
        return
    }

    orderItem, err := h.OrderService.AddOrderItem(ctx, orderID, body.ProductID, body.Quantity)
    if err != nil {
        if err.Error() == "order not found" || err.Error() == "product not found" {
            c.JSON(http.StatusNotFound.Code(), response.ErrorResponse(err.Error(), err.Error()))
            return
        }
        if err.Error() == "insufficient stock" {
            c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("insufficient stock", err.Error()))
            return
        }
        log.ErrorfWithContext(ctx, logTag+" error when adding order item", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when adding order item", err.Error()))
        return
    }

    c.JSON(http.StatusCreated.Code(), gin.H{
        "message":    "order item added successfully",
        "order_item": orderItem,
    })
}

func (h *OrderHandler) UpdateOrderItemHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][UpdateOrderItemHandler]"

    orderID := c.Param("id")
    itemID := c.Param("item_id")

    if orderID == "" || itemID == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("order ID and item ID are required", ""))
        return
    }

    orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid order ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid order ID format", err.Error()))
        return
    }

    itemIDInt, err := strconv.ParseInt(itemID, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid item ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid item ID format", err.Error()))
        return
    }

    var body struct {
        Quantity int32 `json:"quantity" validate:"required,numeric,min=1"`
    }

    if err := c.ShouldBindJSON(&body); err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when deserializing body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid request body", err.Error()))
        return
    }

    if err := validator.ValidateStruct(ctx, body); err.Exists() {
        log.ErrorfWithContext(ctx, logTag+" error when validating the body")
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter valid inputs", err.ErrorMap()))
        return
    }

    updatedItem, err := h.OrderService.UpdateOrderItem(ctx, orderIDInt, itemIDInt, body.Quantity)
    if err != nil {
        if err.Error() == "order not found" || err.Error() == "order item not found" {
            c.JSON(http.StatusNotFound.Code(), response.ErrorResponse(err.Error(), err.Error()))
            return
        }
        if err.Error() == "insufficient stock" {
            c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("insufficient stock", err.Error()))
            return
        }
        log.ErrorfWithContext(ctx, logTag+" error when updating order item", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when updating order item", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message":    "order item updated successfully",
        "order_item": updatedItem,
    })
}

func (h *OrderHandler) RemoveOrderItemHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[OrderHandler][RemoveOrderItemHandler]"

    orderID := c.Param("id")
    itemID := c.Param("item_id")

    if orderID == "" || itemID == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("order ID and item ID are required", ""))
        return
    }

    orderIDInt, err := strconv.ParseInt(orderID, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid order ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid order ID format", err.Error()))
        return
    }

    itemIDInt, err := strconv.ParseInt(itemID, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid item ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid item ID format", err.Error()))
        return
    }

    err = h.OrderService.RemoveOrderItem(ctx, orderIDInt, itemIDInt)
    if err != nil {
        if err.Error() == "order not found" || err.Error() == "order item not found" {
            c.JSON(http.StatusNotFound.Code(), response.ErrorResponse(err.Error(), err.Error()))
            return
        }
        log.ErrorfWithContext(ctx, logTag+" error when removing order item", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when removing order item", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "order item removed successfully",
    })
}