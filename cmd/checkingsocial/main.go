package main

import (
	"checkingsocial/internal/handler"
	"checkingsocial/internal/service"

	"github.com/gin-gonic/gin"
)

func main() {
	// Create a new Gin router
	router := gin.Default()

	// Create the service
	socialService := service.NewSocialChecker()

	// Create the handler
	socialHandler := handler.NewSocialHandler(socialService)

	// Register routes
	socialHandler.RegisterRoutes(router)

	// Start the server
	if err := router.Run(":8080"); err != nil {
		panic(err)
	}
}
