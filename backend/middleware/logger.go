package middleware

import (
	"encoding/json"
	"os"
	"time"

	"github.com/gin-gonic/gin"
)

var jsonLogger = json.NewEncoder(os.Stdout)

type logEntry struct {
	Time       string `json:"time"`
	Level      string `json:"level"`
	Method     string `json:"method"`
	Path       string `json:"path"`
	Query      string `json:"query,omitempty"`
	ClientIP   string `json:"client_ip"`
	Status     int    `json:"status"`
	DurationMs int64  `json:"duration_ms"`
	Error      string `json:"error,omitempty"`
}

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		query := c.Request.URL.RawQuery
		entry := logEntry{
			Time:       start.UTC().Format(time.RFC3339),
			Level:      "info",
			Method:     c.Request.Method,
			Path:       c.Request.URL.Path,
			Query:      query,
			ClientIP:   c.ClientIP(),
			Status:     c.Writer.Status(),
			DurationMs: time.Since(start).Milliseconds(),
		}

		if len(c.Errors) > 0 {
			entry.Level = "error"
			entry.Error = c.Errors.String()
		}

		jsonLogger.Encode(entry)
	}
}
