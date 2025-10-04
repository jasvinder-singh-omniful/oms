package handlers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/log"
	"github.com/omniful/go_commons/validator"
	"github.com/si/internal/storage/service"
	"github.com/si/internal/utils/hash"
	"github.com/si/internal/utils/response"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (h *UserHandler) CreateUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logTag := "[UserHandler][CreateUserHandler]"

	var body struct {
		Name     string `json:"name" validate:"required,alpha"`
		Email    string `json:"email" validate:"required,email"`
		Phone    string `json:"phone" validate:"required,numeric"`
		Password string `json:"password" validate:"required,strong_password"`
	}

	// deserialize JSON -> Struct
	if err := c.ShouldBindJSON(&body); err != nil {
		log.ErrorfWithContext(ctx, logTag+" Failed to bind JSON", err.Error())
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("Invalid request body", err.Error()))
		return
	}

	// validate email
	if validationErr := validator.ValidateStruct(ctx, body); validationErr.Exists() {
		log.ErrorfWithContext(ctx, logTag+" Validation failed", validationErr.ToError())
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("Validation failed", validationErr.Message(), validationErr.ErrorMap()))
		return
	}

	// hash password
	hashedPassword, err := hash.HashPassword(body.Password)
	if err != nil {
		log.ErrorfWithContext(ctx, logTag+" hashing failed", err.Error())
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("error when hashing password ", err.Error()))
		return
	}

	// send user details to service
	createdUser, err := h.UserService.CreateUser(ctx, body.Name, body.Email, body.Phone, hashedPassword)
	if err != nil {
		c.JSON(500, response.ErrorResponse("Failed to create user", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"message": "User created successfully",
		"user":    createdUser,
	})
}

func (h *UserHandler) GetUserByEmailHandler(c *gin.Context){
	ctx := c.Request.Context()

	var body struct {
		Email string `json:"email" validate:"required,email"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter email correctly ", err.Error()))
		return
	}

	if err := validator.ValidateStruct(ctx, body); err.Exists() {
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("validation failed ", err.Message(), err.ErrorMap()))
		return
	}

	user, err := h.UserService.GetUserByEmail(ctx, body.Email)

	if err != nil {
		c.JSON(500, response.ErrorResponse("error when getting user by email ", err.Error()))
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound.Code(), gin.H{
			"message": "user not found",
			"user":    nil,
		})
		return
	}

	c.JSON(http.StatusOK.Code(), gin.H{
		"message": "user fetched successfully",
		"user":    user,
	})
}

func (h *UserHandler) GetUserByIdHandler(c *gin.Context){
	ctx := c.Request.Context()

	var body struct {
		ID int64 `json:"id" validate:"required,numeric"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter id correctly ", err.Error()))
		return
	}

	if err := validator.ValidateStruct(ctx, body); err.Exists() {
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("validation failed ", err.Message(), err.ErrorMap()))
		return
	}

	user, err := h.UserService.GetUserById(ctx, body.ID)

	if err != nil {
		//fmt.Println("error when getting user %v", err.Error())
		c.JSON(500, response.ErrorResponse("error when getting user by id ", err.Error()))
		return
	}

	if user == nil {
		c.JSON(http.StatusNotFound.Code(), gin.H{
			"message": "user not found",
			"user":    nil,
		})
		return
	}

	c.JSON(http.StatusOK.Code(), gin.H{
		"message": "user fetched successfully",
		"user":    user,
	})
}

func (h *UserHandler) UpdateUserHandler(c *gin.Context) {
	ctx := c.Request.Context()
	logTag := "[UserHandler][UpdateUserHandler]"

	var body struct {
		ID       int64  `json:"id" validate:"required,numeric"`
		Name     string `json:"name" validate:"omitempty"`
		Email    string `json:"email" validate:"omitempty,email"`
		Phone    string `json:"phone" validate:"omitempty,numeric"`
		Password string `json:"password" validate:"omitempty,strong_password"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		log.ErrorfWithContext(ctx, logTag+" Failed to bind JSON", err.Error())
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("Invalid request body", err.Error()))
		return
	}

	if validationErr := validator.ValidateStruct(ctx, body); validationErr.Exists() {
		log.ErrorfWithContext(ctx, logTag+" Validation failed", validationErr.ToError())
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("Validation failed", validationErr.Message(), validationErr.ErrorMap()))
		return
	}


	updatedUser, err := h.UserService.UpdateUser(ctx, body.ID, body.Name, body.Email, body.Phone, body.Password)
	if err != nil {
		c.JSON(500, response.ErrorResponse("failed to update user", err.Error()))
		return
	}

	c.JSON(200, gin.H{
		"message": "User updated successfully",
		"user":    updatedUser,
	})
}

func (h *UserHandler) DeleteUserHandler(c *gin.Context) {
	ctx := c.Request.Context()

	var body struct {
		ID int64 `json:"id" validate:"required,numeric"`
	}

	if err := c.ShouldBindJSON(&body); err != nil {
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("please enter id correctly ", err))
		return
	}

	if err := validator.ValidateStruct(ctx, body); err.Exists() {
		c.JSON(http.StatusBadRequest.Code(), response.ErrorResponse("validation failed ", err.Message(), err.ErrorMap()))
		return
	}

	err := h.UserService.DeleteUser(ctx, body.ID)

	fmt.Printf("error delete user: %v\n", err)

	if err != nil {
		c.JSON(500, response.ErrorResponse("error when deleting user ", err.Error()))
		return
	}

	c.JSON(http.StatusOK.Code(), gin.H{
		"message": "user deleted successfully",
	})
}
