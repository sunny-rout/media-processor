package metrics

import (
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
)

var (
	startTime = time.Now()
	mu        sync.RWMutex
	counters  = map[string]*atomic.Int64{}
)

func Inc(key string) {
	mu.Lock()
	c, ok := counters[key]
	if !ok {
		c = &atomic.Int64{}
		counters[key] = c
	}
	mu.Unlock()
	c.Add(1)
}

func snapshot() map[string]int64 {
	mu.RLock()
	defer mu.RUnlock()
	out := make(map[string]int64, len(counters))
	for k, v := range counters {
		out[k] = v.Load()
	}
	return out
}

func Handler(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"uptime_seconds": int64(time.Since(startTime).Seconds()),
		"counters":       snapshot(),
	})
}
