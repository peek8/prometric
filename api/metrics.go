/*
 * Copyright (c) 2025 peek8.io
 *
 * Created Date: Thursday, October 23rd 2025, 4:50:07 pm
 * Author: Md. Asraful Haque
 *
 */

package api

import (
	"bytes"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/v4/process"
)

// ---------- Prometheus metrics ----------
var (
	httpRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests processed, labeled by status code and method.",
		},
		[]string{"code", "method"},
	)

	httpRequestsInProgress = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "http_requests_in_progress",
			Help: "Number of HTTP requests currently being processed.",
		},
		[]string{"method"},
	)

	httpRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Histogram of HTTP request durations in seconds.",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	personStoreCount = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "person_store_count",
		Help: "Number of person records currently stored in memory.",
	})

	personCreatedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "person_created_total",
		Help: "Total number of persons created successfully.",
	})

	personDeletedTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "person_deleted_total",
		Help: "Total number of persons deleted successfully.",
	})

	personNotFoundTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "person_not_found_total",
		Help: "Total number of operations attempted on nonexistent persons.",
	})

	personPayloadSize = prometheus.NewHistogram(prometheus.HistogramOpts{
		Name:    "person_payload_size_bytes",
		Help:    "Size of JSON payloads in POST /person.",
		Buckets: prometheus.ExponentialBuckets(100, 2, 10), // 100B → 50KB
	})


	cpuUsageGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "app_cpu_usage_percent",
		Help: "CPU usage of the Go process (percent).",
	})

	memUsageGauge = prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "app_memory_usage_megabytes",
		Help: "Memory usage of the Go process (MB).",
	})
)

func initMetrics() {
	prometheus.MustRegister(
		httpRequestsTotal, 
		httpRequestsInProgress, 
		httpRequestDuration, 
		personStoreCount, 
		personCreatedTotal,
		personDeletedTotal,
		personNotFoundTotal,
		personPayloadSize,
		cpuUsageGauge,
		memUsageGauge)
	// register default Go runtime metrics (goroutines, memstats)
	//prometheus.MustRegister(prometheus.NewGoCollector())
}

func newResponseRecorder(w http.ResponseWriter) *responseRecorder {
	return &responseRecorder{ResponseWriter: w, statusCode: 200, buf: &bytes.Buffer{}}
}

func (rr *responseRecorder) WriteHeader(code int) {
	rr.statusCode = code
	rr.ResponseWriter.WriteHeader(code)
}

// ---------- System metrics collection ----------
func collectSystemMetricsLoop(stop <-chan struct{}) {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		log.Printf("Failed to create process handle: %v", err)
		return
	}
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-stop:
			return
		case <-ticker.C:
			// Memory RSS bytes -> MB
			if memInfo, err := proc.MemoryInfo(); err == nil {
				memMB := float64(memInfo.RSS) / (1024.0 * 1024.0)
				memUsageGauge.Set(memMB)
			}
			// CPU percent (since last call)
			if cpuPercent, err := proc.CPUPercent(); err == nil {
				cpuUsageGauge.Set(cpuPercent)
			}
			// Also record number of goroutines (optional - via go collector already)
			_ = runtime.NumGoroutine()
		}
	}
}
