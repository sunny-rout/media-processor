package main

import (
	"backend/middleware"
	"backend/routes"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimiter())
	router.Use(middleware.CORS())

	router.GET("/health", routes.HealthCheck)

	api := router.Group("/api")
	{
		api.GET("/validate", routes.ValidateURL)
		api.GET("/stream", routes.StreamMedia)
	}

	log.Println("Starting server on :8080")
	if err := router.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
