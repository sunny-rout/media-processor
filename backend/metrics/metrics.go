package metrics

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var startTime = time.Now()

var (
	registry = prometheus.NewRegistry()

	uptimeSeconds = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "app_uptime_seconds",
		Help: "Application uptime in seconds",
	})

	streamStarted = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "stream_started_total",
		Help: "Total number of streams started",
	})

	streamSuccess = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "stream_success_total",
		Help: "Total number of successful streams",
	})

	streamError = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "stream_errors_total",
		Help: "Total number of stream errors by type",
	}, []string{"type"})

	streamDuration = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "stream_duration_seconds",
		Help:    "Stream duration in seconds",
		Buckets: []float64{1, 5, 10, 30, 60, 120, 300},
	})

	poolSize = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "worker_pool_size",
		Help: "Current worker pool size",
	})
)

func init() {
	registry.MustRegister(uptimeSeconds, streamStarted, streamSuccess, streamError, streamDuration, poolSize)
	go trackUptime()
}

func trackUptime() {
	for range time.Tick(time.Second) {
		uptimeSeconds.Set(time.Since(startTime).Seconds())
	}
}

func Inc(key string) {
	switch key {
	case "stream_started":
		streamStarted.Inc()
	case "stream_success":
		streamSuccess.Inc()
	case "stream_rejected_pool_full":
		streamError.WithLabelValues("pool_full").Inc()
	case "stream_error_private":
		streamError.WithLabelValues("private").Inc()
	case "stream_error_unavailable":
		streamError.WithLabelValues("unavailable").Inc()
	case "stream_error_process":
		streamError.WithLabelValues("process").Inc()
	case "stream_timeout":
		streamError.WithLabelValues("timeout").Inc()
	}
}

func ObserveDuration(seconds float64) {
	streamDuration.Observe(seconds)
}

func SetPoolSize(size int) {
	poolSize.Set(float64(size))
}

// Handler returns the Prometheus metrics handler
func Handler() gin.HandlerFunc {
	h := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	return func(c *gin.Context) {
		h.ServeHTTP(c.Writer, c.Request)
	}
}
