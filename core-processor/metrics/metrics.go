package metrics

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request metrics
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Job processing metrics
	jobsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "jobs_total",
			Help: "Total number of jobs processed",
		},
		[]string{"status", "type"},
	)

	jobDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "job_duration_seconds",
			Help:    "Job processing duration in seconds",
			Buckets: []float64{1, 5, 10, 30, 60, 300, 600, 1800, 3600},
		},
		[]string{"type"},
	)

	// Course generation metrics
	coursesGenerated = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "courses_generated_total",
			Help: "Total number of courses generated",
		},
		[]string{"status"},
	)

	courseGenerationDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "course_generation_duration_seconds",
			Help:    "Course generation duration in seconds",
			Buckets: []float64{30, 60, 120, 300, 600, 900, 1800, 3600},
		},
		[]string{"quality"},
	)

	// System metrics
	activeConnections = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_connections",
			Help: "Number of active connections",
		},
	)

	storageUsage = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "storage_usage_bytes",
			Help: "Storage usage in bytes",
		},
		[]string{"type"},
	)
)

// Init initializes all metrics
func Init() {
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(jobsTotal)
	prometheus.MustRegister(jobDuration)
	prometheus.MustRegister(coursesGenerated)
	prometheus.MustRegister(courseGenerationDuration)
	prometheus.MustRegister(activeConnections)
	prometheus.MustRegister(storageUsage)
}

// Middleware returns Gin middleware for metrics collection
func Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		method := c.Request.Method

		c.Next()

		status := c.Writer.Status()
		duration := time.Since(start).Seconds()

		httpRequestsTotal.WithLabelValues(method, path, string(rune(status))).Inc()
		httpRequestDuration.WithLabelValues(method, path).Observe(duration)
	}
}

// Handler returns the Prometheus metrics handler
func Handler() http.Handler {
	return promhttp.Handler()
}

// RecordJobCompletion records job completion metrics
func RecordJobCompletion(status, jobType string, duration time.Duration) {
	jobsTotal.WithLabelValues(status, jobType).Inc()
	jobDuration.WithLabelValues(jobType).Observe(duration.Seconds())
}

// RecordCourseGeneration records course generation metrics
func RecordCourseGeneration(status, quality string, duration time.Duration) {
	coursesGenerated.WithLabelValues(status).Inc()
	courseGenerationDuration.WithLabelValues(quality).Observe(duration.Seconds())
}

// UpdateActiveConnections updates the active connections gauge
func UpdateActiveConnections(count int) {
	activeConnections.Set(float64(count))
}

// UpdateStorageUsage updates storage usage metrics
func UpdateStorageUsage(storageType string, bytes int64) {
	storageUsage.WithLabelValues(storageType).Set(float64(bytes))
}
