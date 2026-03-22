package middleware

import (
	"log"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		log.Printf("[REQUEST] %s %s from %s", c.Request.Method, c.Request.URL.Path, c.ClientIP())

		c.Next()

		duration := time.Since(startTime)
		statusCode := c.Writer.Status()

		log.Printf("[RESPONSE] %s %s - Status: %d - Duration: %v",
			c.Request.Method,
			c.Request.URL.Path,
			statusCode,
			duration,
		)
	}
}
