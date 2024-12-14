package main

import (
	"log"
	"product-management/database"
	"product-management/models"
	"product-management/routes"

	"github.com/gin-gonic/gin"
)

func main() {
	// Connect to the database
	database.Connect()
	database.ConnectRedis()
	database.ConnectRabbitMQ()
	database.DeclareImageQueue()
	go processImages()

	// Auto-migrate models
	database.DB.AutoMigrate(&models.User{}, &models.Product{})

	// Set up the Gin router
	r := gin.Default()

	// Test API endpoint
	r.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to Product Management API"})
	})

	// Initialize routes
	routes.InitializeRoutes(r)

	// Start the server
	log.Println("Server is running on port 8080")
	r.Run(":8080")

}
