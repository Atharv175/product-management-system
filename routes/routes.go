package routes

import (
	"encoding/json"
	"log"
	"net/http"
	"strconv"

	"product-management/database"
	"product-management/models"

	"github.com/streadway/amqp"

	"github.com/gin-gonic/gin"
)

// InitializeRoutes sets up API endpoints
func InitializeRoutes(router *gin.Engine) {

	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	router.POST("/products", createProduct)
	router.GET("/products/:id", getProductByID)
	router.GET("/products", listProducts)
	router.PUT("/products/:id", updateProduct)
	router.DELETE("/products/:id", deleteProduct)

}

// Create a new product
func createProduct(c *gin.Context) {
	var product models.Product
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Add product to the database
	if err := database.DB.Create(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not create product"})
		return
	}

	// Invalidate Redis cache for the product (in case it already exists in cache)
	cacheKey := strconv.Itoa(int(product.ID))
	err := database.RedisClient.Del(database.Ctx, cacheKey).Err()
	if err != nil {
		log.Println("Failed to invalidate cache for product ID:", product.ID, "Error:", err)
	}

	// Publish image URLs to RabbitMQ for processing
	channel, err := database.GetRabbitMQChannel()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to connect to RabbitMQ"})
		return
	}
	defer channel.Close()

	for _, imageURL := range product.ProductImages {
		err = channel.Publish(
			"",            // Exchange
			"image_queue", // Routing key (queue name)
			false,         // Mandatory
			false,         // Immediate
			amqp.Publishing{
				ContentType: "text/plain",
				Body:        []byte(imageURL),
			},
		)
		if err != nil {
			log.Println("Failed to publish message:", err)
		}
	}

	c.JSON(http.StatusCreated, product)
}

// Get product by ID (with Redis caching)
func getProductByID(c *gin.Context) {
	id := c.Param("id") // Fetch the ID from the path parameter

	// Check Redis cache
	cachedProduct, err := database.RedisClient.Get(database.Ctx, id).Result()
	if err == nil {
		// Cache hit: return cached product
		var product models.Product
		if err := json.Unmarshal([]byte(cachedProduct), &product); err == nil {
			c.JSON(http.StatusOK, product)
			return
		}
	}

	// Cache miss: Fetch from database
	var product models.Product
	if err := database.DB.First(&product, "id = ?", id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Store product in Redis cache
	productJSON, _ := json.Marshal(product)
	database.RedisClient.Set(database.Ctx, id, productJSON, 0) // No expiration

	c.JSON(http.StatusOK, product)
}

// List all products
func listProducts(c *gin.Context) {
	var products []models.Product

	// Parse query parameters
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))    // Default page is 1
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10")) // Default limit is 10
	name := c.Query("name")                                 // Optional filter by name
	minPrice, _ := strconv.ParseFloat(c.Query("min_price"), 64)
	maxPrice, _ := strconv.ParseFloat(c.Query("max_price"), 64)

	// Log the parsed query parameters for debugging
	log.Println("Page:", page, "Limit:", limit, "Name:", name, "Min Price:", minPrice, "Max Price:", maxPrice)

	// Calculate offset
	offset := (page - 1) * limit

	// Build the database query
	query := database.DB.Offset(offset).Limit(limit)

	if name != "" {
		query = query.Where("product_name ILIKE ?", "%"+name+"%") // Case-insensitive search
	}
	if minPrice > 0 {
		query = query.Where("product_price >= ?", minPrice)
	}
	if maxPrice > 0 {
		query = query.Where("product_price <= ?", maxPrice)
	}

	// Execute the query and fetch products
	if err := query.Find(&products).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not retrieve products"})
		return
	}

	// Count the total number of products matching the filters
	var total int64
	countQuery := database.DB.Model(&models.Product{})
	if name != "" {
		countQuery = countQuery.Where("product_name ILIKE ?", "%"+name+"%")
	}
	if minPrice > 0 {
		countQuery = countQuery.Where("product_price >= ?", minPrice)
	}
	if maxPrice > 0 {
		countQuery = countQuery.Where("product_price <= ?", maxPrice)
	}
	countQuery.Count(&total)

	// Return the paginated response
	c.JSON(http.StatusOK, gin.H{
		"page":     page,
		"limit":    limit,
		"total":    total,
		"products": products,
	})
}

func updateProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	// Fetch product by ID from the database
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Bind JSON to the product model
	if err := c.ShouldBindJSON(&product); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Update the product in the database
	if err := database.DB.Save(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not update product"})
		return
	}

	// Invalidate the Redis cache for this product
	cacheKey := id
	err := database.RedisClient.Del(database.Ctx, cacheKey).Err()
	if err != nil {
		log.Println("Failed to invalidate cache for product ID:", id, "Error:", err)
	}

	c.JSON(http.StatusOK, product)
}

func deleteProduct(c *gin.Context) {
	id := c.Param("id")
	var product models.Product

	// Check if the product exists
	if err := database.DB.First(&product, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Product not found"})
		return
	}

	// Delete the product from the database
	if err := database.DB.Delete(&product).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Could not delete product"})
		return
	}

	// Invalidate the Redis cache for the deleted product
	cacheKey := id
	err := database.RedisClient.Del(database.Ctx, cacheKey).Err()
	if err != nil {
		log.Println("Failed to invalidate cache for product ID:", id, "Error:", err)
	}

	c.JSON(http.StatusOK, gin.H{"message": "Product deleted successfully"})
}
