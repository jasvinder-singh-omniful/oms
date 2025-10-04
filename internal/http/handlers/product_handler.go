package handlers

import (
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/validator"
	"github.com/si/internal/storage/service"
	"github.com/si/internal/utils/response"
)


type ProductHandler struct {
	ProductService *service.ProductService
}


func NewProductHandler(productService *service.ProductService) *ProductHandler{
	return &ProductHandler{
		ProductService: productService,
	}
}

func (h *ProductHandler) CreateProductHandler(c *gin.Context) {
	ctx := c.Request.Context()

	logTag := "[ProductHandler][CreateProductHandler]"
	log.InfofWithContext(ctx, logTag+" creating product ")

	var body struct {
		Name          string  `json:"name" validate:"required"`
		SKU           string  `json:"sku" validate:"required,alphanum"`
		Price         float64 `json:"price" validate:"required,numeric"`
		Category      string  `json:"category" validate:"required,alpha"`
		StockQuantity int64   `json:"stock_quantity" validate:"required,numeric"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when deserialsing body")
		c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when creating product", err.Error()))
		return
	}

	if err := validator.ValidateStruct(ctx, body); err.Exists() {
		log.ErrorfWithContext(ctx, logTag+" error when validating the body")
		c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("please enter valid inputs ", err.Error()))
		return
	}

	prod, err := h.ProductService.CreateProduct(ctx, body.Name, body.SKU, body.Price, body.Category, body.StockQuantity)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" error when creating product")
		c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when creaitng product", err.Error()))
		return
	}

	c.JSON(http.StatusCreated.Code(), gin.H{
		"message": "product created successfully",
		"product": prod,
	})
}

func (h *ProductHandler) GetProductByIdHandler(c *gin.Context){
	ctx := c.Request.Context()
	logTag := "[ProductHandler][GetProductByIdHandler]"

	id := c.Param("id")
	if id == ""{
		log.ErrorfWithContext(ctx, logTag+" please enter valid input")
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("product ID is required", ""))
		return
	}

	productId, err := strconv.Atoi(id)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" please enter valid input")
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("product ID is not valid", ""))
		return
	}

	product, err := h.ProductService.GetProductById(ctx, int64(productId))
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when getting product", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when getting product", err.Error()))
        return
    }

	c.JSON(http.StatusOK.Code(), gin.H{
        "message": "product fetched successfully",
        "product": product,
    })
}

func (h *ProductHandler) GetAllProductsHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[ProductHandler][GetAllProductsHandler]"

    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
    
    if page < 1 {
        page = 1
    }
    if limit < 1 || limit > 100 {
        limit = 10
    }

    offset := (page - 1) * limit

    products, total, err := h.ProductService.GetAllProducts(ctx, limit, offset)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when getting products", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when fetching products", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message":  "products fetched successfully",
        "products": products,
        "total":    total,
        "page":     page,
        "limit":    limit,
    })
}

func (h *ProductHandler) SearchProductsHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[ProductHandler][SearchProductsHandler]"

    var body struct {
        Name     string `json:"name" validate:"omitempty"`
        Category string `json:"category" validate:"omitempty"`
        Page     int    `json:"page" validate:"omitempty,numeric"`
        Limit    int    `json:"limit" validate:"omitempty,numeric"`
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

    if body.Page < 1 {
        body.Page = 1
    }
    if body.Limit < 1 || body.Limit > 100 {
        body.Limit = 10
    }

    offset := (body.Page - 1) * body.Limit

    products, total, err := h.ProductService.SearchProduct(ctx, body.Name, body.Category, body.Limit, offset)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when searching products", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when searching products", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message":  "products search completed",
        "products": products,
        "total":    total,
        "page":     body.Page,
        "limit":    body.Limit,
    })
}

func (h *ProductHandler) UpdateProductHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[ProductHandler][UpdateProductHandler]"

    // Get ID from path parameter
    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("product ID is required", ""))
        return
    }
	
    productID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid product ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid product ID format", err.Error()))
        return
    }

    var body struct {
        Name          string  `json:"name" validate:"omitempty"`
        Price         float64 `json:"price" validate:"omitempty,numeric"`
        Category      string  `json:"category" validate:"omitempty,alpha"`
        StockQuantity int64   `json:"stock_quantity" validate:"omitempty,numeric"`
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

    updatedProduct, err := h.ProductService.UpdateProduct(ctx, productID, body.Name, body.Price, body.Category, body.StockQuantity)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when updating product", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when updating product", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "product updated successfully",
        "product": updatedProduct,
    })
}

func (h *ProductHandler) DeleteProductHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[ProductHandler][DeleteProductHandler]"

    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("product ID is required", ""))
        return
    }

    productID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid product ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid product ID format", err.Error()))
        return
    }

    err = h.ProductService.DeleteProduct(ctx, productID)
    if err != nil {
        if err.Error() == "product not found" {
            c.JSON(http.StatusNotFound.Code(), response.ErrorResponse("product not found", err.Error()))
            return
        }
        log.ErrorfWithContext(ctx, logTag+" error when deleting product", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when deleting product", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "product deleted successfully",
    })
}

func (h *ProductHandler) UpdateInventoryHandler(c *gin.Context) {
    ctx := c.Request.Context()
    logTag := "[ProductHandler][UpdateInventoryHandler]"

    id := c.Param("id")
    if id == "" {
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("product ID is required", ""))
        return
    }

    productID, err := strconv.ParseInt(id, 10, 64)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" invalid product ID format", err)
        c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("invalid product ID format", err.Error()))
        return
    }

    var body struct {
        StockQuantity int64  `json:"stock_quantity" validate:"required,numeric"`
        Operation     string `json:"operation" validate:"required,oneof=set add subtract"`
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

    updatedProduct, err := h.ProductService.UpdateInventory(ctx, productID, body.StockQuantity, body.Operation)
    if err != nil {
        log.ErrorfWithContext(ctx, logTag+" error when updating inventory", err)
        c.JSON(http.StatusInternalServerError.Code(), response.ErrorResponse("error when updating inventory", err.Error()))
        return
    }

    c.JSON(http.StatusOK.Code(), gin.H{
        "message": "inventory updated successfully",
        "product": updatedProduct,
    })
}

