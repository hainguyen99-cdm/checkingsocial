package main

import (
	"checkingsocial/internal/handler"
	"checkingsocial/internal/service"
	"checkingsocial/pkg/cache"
	"checkingsocial/pkg/cronjob"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables from .env file
	_ = godotenv.Load()

	// Initialize Redis
	if err := cache.InitRedis(); err != nil {
		log.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer cache.Close()

	// Initialize cronjob scheduler
	if err := cronjob.InitCronScheduler(); err != nil {
		log.Fatalf("Failed to initialize cronjob scheduler: %v", err)
	}
	defer cronjob.StopCronScheduler()

	// Fetch followers immediately on startup
	if err := cronjob.FetchFollowersNow(); err != nil {
		log.Printf("Warning: Failed to fetch followers on startup: %v", err)
	}

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

	// Handle graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		<-sigChan
		log.Println("Shutting down server...")
		cronjob.StopCronScheduler()
		cache.Close()
		os.Exit(0)
	}()

	// Start the server
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
