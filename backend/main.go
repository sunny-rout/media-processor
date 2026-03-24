package main

import (
	"backend/metrics"
	"backend/middleware"
	"backend/routes"
	"context"
	"encoding/json"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
)

var startupLog = json.NewEncoder(os.Stdout)

type startLog struct {
	Time    string `json:"time"`
	Level   string `json:"level"`
	Message string `json:"msg"`
}

func logStartup(msg string) {
	startupLog.Encode(startLog{
		Time:    time.Now().UTC().Format(time.RFC3339),
		Level:   "info",
		Message: msg,
	})
}

func main() {
	router := gin.New()

	router.Use(gin.Recovery())
	router.Use(middleware.Logger())
	router.Use(middleware.RateLimiter())
	router.Use(middleware.CORS())

	router.GET("/health", routes.HealthCheck)
	router.GET("/metrics", metrics.Handler)

	api := router.Group("/api")
	{
		api.GET("/validate", routes.ValidateURL)
		api.GET("/stream", routes.StreamMedia)
	}

	srv := &http.Server{
		Addr:         ":8080",
		Handler:      router,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 4 * time.Minute,
		IdleTimeout:  60 * time.Second,
	}

	go func() {
		logStartup("server starting on :8080")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			startupLog.Encode(startLog{
				Time:    time.Now().UTC().Format(time.RFC3339),
				Level:   "fatal",
				Message: "server failed: " + err.Error(),
			})
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logStartup("shutdown signal received, draining connections")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		startupLog.Encode(startLog{
			Time:    time.Now().UTC().Format(time.RFC3339),
			Level:   "error",
			Message: "forced shutdown: " + err.Error(),
		})
		os.Exit(1)
	}

	logStartup("server stopped cleanly")
}
