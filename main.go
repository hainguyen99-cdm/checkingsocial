package main

import (
	"checkingsocial/internal/handler"
	"checkingsocial/internal/service"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// Khởi tạo router của Gin
	router := gin.Default()

	// Dependency Injection: Tạo các instance
	socialCheckerService := service.NewSocialChecker()
	socialHandler := handler.NewSocialHandler(socialCheckerService)

	// Đăng ký các routes
	socialHandler.RegisterRoutes(router)

	// Cấu hình địa chỉ và cổng cho server
	serverAddr := ":8080"
	log.Printf("Server is running on %s", serverAddr)

	// Khởi chạy server
	if err := router.Run(serverAddr); err != nil {
		log.Fatalf("Failed to run server: %v", err)
	}
}
