package main

import (
	"log"
	"os"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	config "github.com/phillip/contribution-tracker-go/config"
	routes "github.com/phillip/contribution-tracker-go/routes"
)

func main() {
	 // Load env
    if err := godotenv.Load(); err != nil {
        log.Println("no .env file loaded, reading environment variables")
    }

    // Load app config (JWT secret, etc.)
    cfg, err := config.LoadConfig()
    if err != nil {
        log.Fatalf("config load error: %v", err)
    }

    // ✅ Connect to MongoDB first
    client := config.ConnectDB()
    if client == nil {
        log.Fatal("❌ Could not connect to MongoDB")
    }
    log.Println("✅ Connected to MongoDB")

    // ✅ Now ensure indexes
    config.EnsureAllIndexes(client, cfg.DBName)

	// Gin router
	r := gin.Default()

	// CORS configuration
	r.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"https://contribution-khaki.vercel.app",
			"http://localhost:4200",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders: []string{"Origin", "Content-Type", "Authorization", "If-None-Match", "If-Modified-Since",
		},
		ExposeHeaders:    []string{"ETag", "Last-Modified", "Content-Length"}, 
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))


	routes.SetupRoutes(r, cfg)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("🚀 Listening on :%s\n", port)
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
