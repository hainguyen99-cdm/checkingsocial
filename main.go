package main

import (
	"checkingsocial/internal/handler"
	"checkingsocial/internal/service"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)

	// Use gin.New() for a clean router, then add middleware manually
	router := gin.New()
	router.Use(gin.Recovery()) // Add recovery middleware to catch panics

	// Dependency Injection: Create instances
	socialCheckerService := service.NewSocialChecker()
	socialHandler := handler.NewSocialHandler(socialCheckerService)

	// Register routes
	socialHandler.RegisterRoutes(router)

	// Configure server address and port
	serverAddr := ":8080"
	log.Printf("Server is running on %s", serverAddr)

	// Start the server
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
