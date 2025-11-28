package main

import (
	"net/http"
	"os"
	"polcompass/backend/models"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	url := os.Getenv("DATABASE_URL")
	isLocalMode := os.Getenv("LOCAL_MODE")

	time.Sleep(2 * time.Second)

	dsn := url
	if !strings.Contains(dsn, "sslmode=") && isLocalMode != "true" {
		dsn += "?sslmode=require"
	}
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})

	if err != nil {
		panic("failed to connect database")
	}

	// Migrate the schema
	db.AutoMigrate(&models.Polcompass{})
	db.AutoMigrate(&models.Question{})

	router := gin.Default()

	router.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	// Add CORS middleware
	router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "*"}, // Replace with your frontend domains
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
	}))

	polCompassController := &models.PolCompassController{DB: db}

	router.POST("/polcompass", polCompassController.POST)

	router.GET("/polcompass", polCompassController.GET)

	router.GET("/polcompass/first", polCompassController.First)

	router.GET("/summary", polCompassController.Summary)

	router.Run() // listen and serve on 0.0.0.0:8080
}
