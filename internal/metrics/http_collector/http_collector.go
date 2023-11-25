package httpcollector

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/todesdev/go-obs/internal/logging"
)

const (
	HttpSubsystem = "http"

	HttpRequestsTotal           = "requests_total"
	HttpRequestDurationSeconds  = "request_duration_seconds"
	HttpRequestsInProgressTotal = "requests_in_progress_total"

	HttpRequestsHelp                = "Total number of HTTP requests."
	HttpRequestsDurationSecondsHelp = "Duration of HTTP requests."
	HttpRequestsInProgressHelp      = "Number of HTTP requests in progress."

	HttpStatusCodeLabel = "status_code"
	HttpMethodLabel     = "method"
)

var (
	httpCollector *HttpCollector
)

type HttpCollector struct {
	requestCount     *prometheus.CounterVec
	responseTime     *prometheus.HistogramVec
	requestsInFlight *prometheus.GaugeVec
}

func newHttpCollector(serviceName string) *HttpCollector {
	requestCount := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: prometheus.BuildFQName(serviceName, HttpSubsystem, HttpRequestsTotal),
			Help: HttpRequestsHelp,
		},
		[]string{HttpMethodLabel, HttpStatusCodeLabel},
	)

	responseTime := prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    prometheus.BuildFQName(serviceName, HttpSubsystem, HttpRequestDurationSeconds),
			Help:    HttpRequestsDurationSecondsHelp,
			Buckets: prometheus.DefBuckets,
		},
		[]string{HttpMethodLabel, HttpStatusCodeLabel},
	)

	requestsInFlight := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: prometheus.BuildFQName(serviceName, HttpSubsystem, HttpRequestsInProgressTotal),
			Help: HttpRequestsInProgressHelp,
		},
		[]string{HttpMethodLabel},
	)

	httpCollector = &HttpCollector{
		requestCount:     requestCount,
		responseTime:     responseTime,
		requestsInFlight: requestsInFlight,
	}

	return httpCollector
}

func (collector *HttpCollector) Register(registry *prometheus.Registry) {
	registry.MustRegister(collector.requestCount)
	registry.MustRegister(collector.responseTime)
	registry.MustRegister(collector.requestsInFlight)
}

func Setup(registry *prometheus.Registry, serviceName string) {
	logger := logging.LoggerWithProcess("HttpCollectorSetup")
	logger.Info("Setting up HTTP metrics...")
	newHttpCollector(serviceName).Register(registry)

	logger.Info("HTTP metrics setup complete")
}

func GetHttpCollector() *HttpCollector {
	return httpCollector
}

func (collector *HttpCollector) IncRequestCount(method string, statusCode int) {

	collector.requestCount.WithLabelValues(method, strconv.Itoa(statusCode)).Inc()
}

func (collector *HttpCollector) ObserveResponseTime(method string, statusCode int, duration time.Duration) {
	collector.responseTime.WithLabelValues(method, strconv.Itoa(statusCode)).Observe(float64(duration) / float64(time.Second))
}

func (collector *HttpCollector) IncRequestsInFlight(method string) {
	collector.requestsInFlight.WithLabelValues(method).Inc()
}

func (collector *HttpCollector) DecRequestsInFlight(method string) {
	collector.requestsInFlight.WithLabelValues(method).Dec()
}
