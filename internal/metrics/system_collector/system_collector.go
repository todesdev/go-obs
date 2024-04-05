package systemcollector

import (
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/todesdev/go-obs/internal/logging"
)

const (
	SystemSubsystem      = "system"
	SystemGCStats        = "gc_stats"
	SystemGoRoutineCount = "go_routine_count"

	SystemGCStatsHelp        = "GC stats."
	SystemGoRoutineCountHelp = "Number of go routines."
)

type MetricsCollector struct {
	gcStatsDesc        *prometheus.Desc
	goRoutineCountDesc *prometheus.Desc
}

func newCollector(serviceName string) *MetricsCollector {
	collector := &MetricsCollector{
		gcStatsDesc: prometheus.NewDesc(
			prometheus.BuildFQName(serviceName, SystemSubsystem, SystemGCStats),
			SystemGCStatsHelp,
			nil, nil,
		),
		goRoutineCountDesc: prometheus.NewDesc(
			prometheus.BuildFQName(serviceName, SystemSubsystem, SystemGoRoutineCount),
			SystemGoRoutineCountHelp,
			nil, nil,
		),
	}

	return collector
}

func (c *MetricsCollector) Collect(ch chan<- prometheus.Metric) {
	// GC Stats
	var gcStats runtime.MemStats
	runtime.ReadMemStats(&gcStats)
	ch <- prometheus.MustNewConstMetric(c.gcStatsDesc, prometheus.GaugeValue, float64(gcStats.PauseTotalNs)/1e9)

	// Goroutines Count
	ch <- prometheus.MustNewConstMetric(c.goRoutineCountDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
}

func (c *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.gcStatsDesc
	ch <- c.goRoutineCountDesc
}

func Setup(registry *prometheus.Registry, serviceName string) {
	logger := logging.LoggerWithProcess("MetricsSystemCollector")
	logger.Info("Setting up system metrics...")

	registry.MustRegister(newCollector(serviceName))

	logger.Info("System metrics setup complete")
}
