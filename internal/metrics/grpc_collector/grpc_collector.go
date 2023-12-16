package grpc_collector

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/todesdev/go-obs/internal/logging"
	"sync"
	"time"
)

const (
	GrpcSubsystem = "grpc"

	GrpcRequestsTotal           = "requests_total"
	GrpcRequestDurationSeconds  = "request_duration_seconds"
	GrpcRequestsInProgressTotal = "requests_in_progress_total"

	GrpcRequestsHelp                = "Total number of gRPC requests."
	GrpcRequestDurationSecondsHelp = "Duration of gRPC requests."
	GrpcRequestsInProgressHelp      = "Number of gRPC requests in progress."

	GrpcStatusCodeLabel = "status_code"
	GrpcMethodLabel     = "method"
)

var (
	grpcCollector *GrpcCollector
)

type GrpcCollector struct {
	mu               sync.Mutex
	requestCount     *prometheus.CounterVec
	responseTime     *prometheus.HistogramVec
	requestsInFlight *prometheus.GaugeVec
}

func newGrpcCollector(serviceName string) *GrpcCollector {
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(serviceName, GrpcSubsystem, GrpcRequestsTotal),
			Help: GrpcRequestsHelp,
		},
		[]string{GrpcMethodLabel, GrpcStatusCodeLabel},
	)

	responseTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prometheus.BuildFQName(serviceName, GrpcSubsystem, GrpcRequestDurationSeconds),
			Help:    GrpcRequestDurationSecondsHelp,
			Buckets: prometheus.DefBuckets,
		},
		[]string{GrpcMethodLabel, GrpcStatusCodeLabel},
	)

	requestsInFlight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(serviceName, GrpcSubsystem, GrpcRequestsInProgressTotal),
			Help: GrpcRequestsInProgressHelp,
		},
		[]string{GrpcMethodLabel},
	)

	grpcCollector = &GrpcCollector{
		requestCount:     requestCount,
		responseTime:     responseTime,
		requestsInFlight: requestsInFlight,
	}

	return grpcCollector
}

func (collector *GrpcCollector) Register(registry *prometheus.Registry) {
	registry.MustRegister(collector.requestCount)
	registry.MustRegister(collector.responseTime)
	registry.MustRegister(collector.requestsInFlight)
}

func Setup(registry *prometheus.Registry, serviceName string) {
	logger := logging.LoggerWithProcess("GrpcCollectorSetup")
	logger.Info("Setting up gRPC metrics...")
	newGrpcCollector(serviceName).Register(registry)

	logger.Info("gRPC metrics setup complete")
}

func GetGrpcCollector() *GrpcCollector {
	return grpcCollector
}

func (collector *GrpcCollector) IncRequestCount(method, statusCode string) {
	collector.mu.Lock()
	collector.requestCount.WithLabelValues(method, statusCode).Inc()
	collector.mu.Unlock()
}

func (collector *GrpcCollector) ObserveResponseTime(method, statusCode string, duration time.Duration) {
	collector.mu.Lock()
	collector.responseTime.WithLabelValues(method, statusCode).Observe(float64(duration) / float64(time.Second))
	collector.mu.Unlock()
}

func (collector *GrpcCollector) IncRequestsInFlight(method string) {
	collector.mu.Lock()
	collector.requestsInFlight.WithLabelValues(method).Inc()
	collector.mu.Unlock()
}

func (collector *GrpcCollector) DecRequestsInFlight(method string) {
	collector.mu.Lock()
	collector.requestsInFlight.WithLabelValues(method).Dec()
	collector.mu.Unlock()
}