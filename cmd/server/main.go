package main

import (
	"context"
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/omniful/go_commons/http"
	"github.com/omniful/go_commons/log"
	"github.com/si/internal/storage/postgres"
	"github.com/si/internal/storage/service"
	"github.com/si/internal/types"
	user_handler "github.com/si/internal/http/handlers/user_handler"
	"github.com/si/internal/config"
	//"github.com/si/internal/utils/response"
)

func main() {
	// initialize all the configs
	ctx := context.Background()

	err := config.Init(ctx)
	if err != nil {
		panic(fmt.Errorf("failed to initialize configuration: %w", err))
	}

	// initializing the logger
	dev_env := config.AppConf.Environment == "local"

	logLevel := "info"
	log_format := "json"

	if dev_env {
		logLevel = "debug"
		log_format = "text"
	}

	err = log.InitializeLogger(
		log.Formatter(log_format),
		log.Level(logLevel),
		log.ColoredLevelEncoder(),
	)
	if err != nil {
		log.Error("Logger init error")
	}

	log.Info("initialization done")


	// initializing storage-Postgres DB
	cluster := postgres.NewPostgres(ctx)
	log.Info("database initialization done")

	// repos
	userRepo := postgres.NewUserRepo(ctx, cluster)

	// services
	userService := service.NewUserService(userRepo)

	// handlers
	userHandler := user_handler.NewUserHandler(userService)


	server := http.InitializeServer(
		":3000", 0, 0, 0, true,
	)

	server.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK.Code(), types.APIResponse{
			Message: "service is healthy",
			Data:    map[string]string{"status": "healthy"},
		})
	})


	server.POST("/users", userHandler.CreateUserHandler)
	server.POST("/users/email", userHandler.GetUserByEmailHandler)
	server.POST("/users/id", userHandler.GetUserByIdHandler)
	server.PUT("/users", userHandler.UpdateUserHandler)
	server.DELETE("/users", userHandler.DeleteUserHandler)

	log.Info("server starting on port 3000")
	if err := server.StartServer("oms-service"); err != nil {
		log.Error("error while starting the server")
	}
}
