package main

import (
	"expense_tracker/internal/database"
	"expense_tracker/internal/handlers"
	"log"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
)

func main() {
	// Ensure data directories exist
	createRequiredDirectories()

	// Initialize database
	dbPath := getEnv("DB_PATH", "data/expense_tracker.db")
	db := database.InitDB(dbPath)
	defer db.Close()

	// Create handlers
	tagHandler := handlers.NewTagHandler(db)
	paymentHandler := handlers.NewPaymentHandler(db)
	documentHandler := handlers.NewDocumentHandler(db)

	// Setup router
	router := gin.Default()

	// Add CORS middleware
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API routes
	api := router.Group("/api")
	{
		// Health check
		api.GET("/health", func(c *gin.Context) {
			c.JSON(200, gin.H{"status": "ok"})
		})

		// Register handlers
		tagHandler.RegisterRoutes(api)
		paymentHandler.RegisterRoutes(api)
		documentHandler.RegisterRoutes(api)
	}

	// Static file serving for frontend
	router.Static("/static", "./static")
	router.NoRoute(func(c *gin.Context) {
		c.File("./static/index.html")
	})

	// Start server
	port := getEnv("PORT", "8080")
	log.Printf("Server starting on port %s", port)
	if err := router.Run(":" + port); err != nil {
		log.Fatal(err)
	}
}

func createRequiredDirectories() {
	dirs := []string{
		"data",
		"storage",
		filepath.Join("storage", "documents"),
		filepath.Join("storage", "invoices"),
		"static",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			log.Fatalf("Failed to create directory %s: %v", dir, err)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
