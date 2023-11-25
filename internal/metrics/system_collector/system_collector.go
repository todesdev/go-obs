package systemcollector

import (
	"os"
	"runtime"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"github.com/todesdev/go-obs/internal/logging"
)

const (
	SystemSubsystem        = "system"
	SystemCPUUsagePercent  = "cpu_usage_percent"
	SystemMemoryUsageBytes = "memory_usage_bytes"
	SystemMemoryTotalBytes = "memory_total_bytes"
	SystemGCStats          = "gc_stats"
	SystemGoRoutineCount   = "go_routine_count"

	SystemCpuUsagePercentHelp  = "CPU usage as a percentage."
	SystemMemoryUsageBytesHelp = "Memory usage in bytes."
	SystemMemoryTotalBytesHelp = "Total memory in bytes."
	SystemGCStatsHelp          = "GC stats."
	SystemGoRoutineCountHelp   = "Number of go routines."
)

type MetricsCollector struct {
	proc               *process.Process
	cpuUsageDesc       *prometheus.Desc
	memoryUsageDesc    *prometheus.Desc
	memoryTotalDesc    *prometheus.Desc
	gcStatsDesc        *prometheus.Desc
	goRoutineCountDesc *prometheus.Desc
}

func newCollector(serviceName string) *MetricsCollector {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		panic(err)
	}

	collector := &MetricsCollector{
		proc: proc,
		cpuUsageDesc: prometheus.NewDesc(
			prometheus.BuildFQName(serviceName, SystemSubsystem, SystemCPUUsagePercent),
			SystemCpuUsagePercentHelp,
			nil, nil,
		),
		memoryUsageDesc: prometheus.NewDesc(
			prometheus.BuildFQName(serviceName, SystemSubsystem, SystemMemoryUsageBytes),
			SystemMemoryUsageBytesHelp,
			nil, nil,
		),
		memoryTotalDesc: prometheus.NewDesc(
			prometheus.BuildFQName(serviceName, SystemSubsystem, SystemMemoryTotalBytes),
			SystemMemoryTotalBytesHelp,
			nil, nil,
		),
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
	// CPU Usage
	cpuPercent, _ := c.proc.CPUPercent()
	ch <- prometheus.MustNewConstMetric(c.cpuUsageDesc, prometheus.GaugeValue, cpuPercent)

	// Memory Usage
	memInfo, _ := c.proc.MemoryInfo()
	ch <- prometheus.MustNewConstMetric(c.memoryUsageDesc, prometheus.GaugeValue, float64(memInfo.RSS))

	// System Total Memory
	vmStat, _ := mem.VirtualMemory()
	ch <- prometheus.MustNewConstMetric(c.memoryTotalDesc, prometheus.GaugeValue, float64(vmStat.Total))

	// GC Stats
	var gcStats runtime.MemStats
	runtime.ReadMemStats(&gcStats)
	ch <- prometheus.MustNewConstMetric(c.gcStatsDesc, prometheus.GaugeValue, float64(gcStats.PauseTotalNs)/1e9)

	// Goroutines Count
	ch <- prometheus.MustNewConstMetric(c.goRoutineCountDesc, prometheus.GaugeValue, float64(runtime.NumGoroutine()))
}

func (c *MetricsCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- c.cpuUsageDesc
	ch <- c.memoryUsageDesc
	ch <- c.memoryTotalDesc
	ch <- c.gcStatsDesc
	ch <- c.goRoutineCountDesc
}

func Setup(registry *prometheus.Registry, serviceName string) {
	logger := logging.LoggerWithProcess("MetricsSystemCollector")
	logger.Info("Setting up system metrics...")

	registry.MustRegister(newCollector(serviceName))

	logger.Info("System metrics setup complete")
}
