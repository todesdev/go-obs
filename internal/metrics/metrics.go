package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/todesdev/go-obs/internal/logging"
	httpcollector "github.com/todesdev/go-obs/internal/metrics/http_collector"
	natscollector "github.com/todesdev/go-obs/internal/metrics/nats_collector"
	systemcollector "github.com/todesdev/go-obs/internal/metrics/system_collector"
)

func Setup(serviceName string) *prometheus.Registry {
	logger := logging.LoggerWithProcess("MetricsSetup")
	logger.Info("Setting up metrics...")

	registry := prometheus.NewRegistry()

	systemcollector.Setup(registry, serviceName)
	httpcollector.Setup(registry, serviceName)
	natscollector.Setup(registry, serviceName)

	logger.Info("Metrics setup complete")

	return registry
}
