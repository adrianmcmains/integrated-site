package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/spf13/viper"
)

func main() {
	// Load configuration
	loadConfig()

	// Connect to database
	dbPool, err := connectDB()
	if err != nil {
		log.Fatalf("Unable to connect to database: %v\n", err)
	}
	defer dbPool.Close()

	// Initialize router
	router := setupRouter(dbPool)

	// Start server
	server := &http.Server{
		Addr:    fmt.Sprintf(":%s", viper.GetString("server.port")),
		Handler: router,
	}

	// Start server in a goroutine so it doesn't block graceful shutdown
	go func() {
		log.Printf("Server running on port %s\n", viper.GetString("server.port"))
		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Error starting server: %v\n", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Create a context with timeout for shutdown
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v\n", err)
	}

	log.Println("Server exited properly")
}

func loadConfig() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	viper.SetDefault("server.port", "8080")
	viper.SetDefault("database.host", "localhost")
	viper.SetDefault("database.port", "5432")
	viper.SetDefault("database.name", "integrated_site")
	viper.SetDefault("database.user", "postgres")
	viper.SetDefault("database.sslmode", "disable")

	if err := viper.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Println("No config file found, using defaults")
		} else {
			log.Fatalf("Error reading config file: %v\n", err)
		}
	}
}

func connectDB() (*pgxpool.Pool, error) {
	connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		viper.GetString("database.host"),
		viper.GetString("database.port"),
		viper.GetString("database.user"),
		viper.GetString("database.password"),
		viper.GetString("database.name"),
		viper.GetString("database.sslmode"),
	)

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, err
	}

	pool, err := pgxpool.ConnectConfig(context.Background(), config)
	if err != nil {
		return nil, err
	}

	// Test connection
	if err := pool.Ping(context.Background()); err != nil {
		return nil, err
	}

	return pool, nil
}

func setupRouter(dbPool *pgxpool.Pool) *gin.Engine {
	router := gin.Default()

	// Middleware
	router.Use(gin.Logger())
	router.Use(gin.Recovery())
	
	// Set up CORS
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Health check
	router.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status": "ok",
			"time":   time.Now().Format(time.RFC3339),
		})
	})

	// API routes
	api := router.Group("/api")
	{
		// Blog routes
		blog := api.Group("/blog")
		{
			blog.GET("/posts", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get all posts"})
			})
			blog.GET("/posts/:slug", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get post by slug"})
			})
			blog.GET("/categories", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get all categories"})
			})
			blog.GET("/tags", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get all tags"})
			})
		}

		// Shop routes
		shop := api.Group("/shop")
		{
			shop.GET("/products", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get all products"})
			})
			shop.GET("/products/:slug", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get product by slug"})
			})
			shop.GET("/categories", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get all product categories"})
			})
		}

		// Order routes
		orders := api.Group("/orders")
		{
			orders.POST("/", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Create new order"})
			})
			orders.GET("/:id", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get order by ID"})
			})
		}

		// Auth routes
		auth := api.Group("/auth")
		{
			auth.POST("/register", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Register new user"})
			})
			auth.POST("/login", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Login user"})
			})
			auth.GET("/profile", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get user profile"})
			})
		}

		// CMS routes
		cms := api.Group("/cms")
		{
			cms.GET("/pages", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get all pages"})
			})
			cms.GET("/pages/:slug", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Get page by slug"})
			})
		}

		// Payment routes
		payment := api.Group("/payment")
		{
			payment.POST("/eversend/init", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Initialize payment with Eversend"})
			})
			payment.POST("/eversend/webhook", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{"message": "Eversend webhook handler"})
			})
		}
	}

	// Admin routes (protected)
	admin := router.Group("/admin")
	{
		admin.GET("/dashboard", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"message": "Admin dashboard data"})
		})
	}

	return router
}