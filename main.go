package main

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTP request count metric
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	// Request latency histogram
	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "endpoint"},
	)

	// Application uptime metric
	appStartTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "app_start_time_seconds",
			Help: "Application start time in Unix timestamp",
		},
	)

	// Custom business metric - processed jobs total
	processedJobsTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "processed_jobs_total",
			Help: "Total number of jobs processed",
		},
	)

	// Custom business metric - active jobs gauge
	activeJobsGauge = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_jobs",
			Help: "Number of currently active jobs",
		},
	)
)

func init() {
	// Register metrics with Prometheus
	prometheus.MustRegister(httpRequestsTotal)
	prometheus.MustRegister(httpRequestDuration)
	prometheus.MustRegister(appStartTime)
	prometheus.MustRegister(processedJobsTotal)
	prometheus.MustRegister(activeJobsGauge)

	// Set application start time
	appStartTime.Set(float64(time.Now().Unix()))
}

func main() {
	port := "8080"
	if portEnv := os.Getenv("PORT"); portEnv != "" {
		port = portEnv
	}

	// Main application endpoint
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		// Simulate some work
		processJob()
		
		duration := time.Since(start).Seconds()
		status := "200"
		
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, status).Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Hello from Prometheus Demo App!\n")
		fmt.Fprintf(w, "Uptime: %s\n", time.Since(time.Unix(int64(appStartTime.Get()), 0)).Round(time.Second))
		fmt.Fprintf(w, "Processed Jobs: %d\n", int(processedJobsTotal.Get()))
		fmt.Fprintf(w, "Active Jobs: %d\n", int(activeJobsGauge.Get()))
	})

	// Health endpoint
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "OK")
		
		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})

	// Readiness endpoint
	http.HandleFunc("/ready", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		
		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, "Ready")
		
		duration := time.Since(start).Seconds()
		httpRequestsTotal.WithLabelValues(r.Method, r.URL.Path, "200").Inc()
		httpRequestDuration.WithLabelValues(r.Method, r.URL.Path).Observe(duration)
	})

	// Metrics endpoint
	http.Handle("/metrics", promhttp.Handler())

	fmt.Printf("Starting server on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		fmt.Printf("Error starting server: %v\n", err)
		os.Exit(1)
	}
}

// processJob simulates processing a job
func processJob() {
	processedJobsTotal.Inc()
	
	// Simulate active job
	activeJobsGauge.Inc()
	time.Sleep(10 * time.Millisecond)
	activeJobsGauge.Dec()
}