package users

import (
	"github.com/gin-gonic/gin"
	"github.com/si/internal/storage/service"
	"github.com/si/internal/types"
)

type UserHandler struct {
	UserService *service.UserService
}

func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		UserService: userService,
	}
}

func (h *UserHandler) CreateUser(c *gin.Context)  {
	ctx := c.Request.Context()

	var user *types.User
	h.UserService.CreateUser(ctx, user)

	c.JSON(200, gin.H{
		"message": "User created successfully",
	})
}

