package system_metrics

import (
	"context"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/process"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
	"os"
	"runtime"
)

const (
	SystemSubsystem            = "system"
	SystemCPUUsagePercent      = "cpu_usage_percent"
	SystemCPUUsagePercentHelp  = "CPU usage as a percentage."
	SystemMemoryUsageBytes     = "memory_usage_bytes"
	SystemMemoryUsageBytesHelp = "Memory usage in bytes."
	SystemMemoryTotalBytes     = "memory_total_bytes"
	SystemMemoryTotalBytesHelp = "Total memory in bytes."
	SystemGCStats              = "gc_stats"
	SystemGCStatsHelp          = "Garbage collection statistics."
	SystemGoRoutineCount       = "go_routine_count"
	SystemGoRoutineCountHelp   = "Number of go routines."
)

var (
	meter = otel.Meter(SystemSubsystem)
)

func Setup() error {
	proc, err := process.NewProcess(int32(os.Getpid()))
	if err != nil {
		return err
	}

	if err := setupCpuUsagePercent(proc); err != nil {
		return err
	}

	if err := setupMemoryUsageBytes(proc); err != nil {
		return err
	}

	if err := setupMemoryTotalBytes(); err != nil {
		return err
	}

	if err := setupGcStats(); err != nil {
		return err
	}

	if err := setupGoRoutineCount(); err != nil {
		return err
	}

	return nil

}

func setupCpuUsagePercent(proc *process.Process) error {
	if _, err := meter.Float64ObservableGauge(
		SystemCPUUsagePercent,
		metric.WithDescription(SystemCPUUsagePercentHelp),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			cpuPercent, err := proc.CPUPercent()
			if err != nil {
				return err
			}

			o.Observe(cpuPercent)
			return nil
		}),
	); err != nil {
		return err
	}

	return nil
}

func setupMemoryUsageBytes(proc *process.Process) error {
	if _, err := meter.Float64ObservableGauge(
		SystemMemoryUsageBytes,
		metric.WithDescription(SystemMemoryUsageBytesHelp),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			memInfo, err := proc.MemoryInfo()
			if err != nil {
				return err
			}

			o.Observe(float64(memInfo.RSS))
			return nil
		}),
	); err != nil {
		return err
	}

	return nil
}

func setupMemoryTotalBytes() error {
	if _, err := meter.Float64ObservableGauge(
		SystemMemoryTotalBytes,
		metric.WithDescription(SystemMemoryTotalBytesHelp),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			vmStat, err := mem.VirtualMemory()
			if err != nil {
				return err
			}

			o.Observe(float64(vmStat.Total))
			return nil
		}),
	); err != nil {
		return err
	}

	return nil
}

func setupGcStats() error {
	if _, err := meter.Float64ObservableGauge(
		SystemGCStats,
		metric.WithDescription(SystemGCStatsHelp),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			var gcStats runtime.MemStats
			runtime.ReadMemStats(&gcStats)

			o.Observe(float64(gcStats.PauseTotalNs) / 1e9)
			return nil
		}),
	); err != nil {
		return err
	}

	return nil
}

func setupGoRoutineCount() error {
	if _, err := meter.Float64ObservableGauge(
		SystemGoRoutineCount,
		metric.WithDescription(SystemGoRoutineCountHelp),
		metric.WithFloat64Callback(func(_ context.Context, o metric.Float64Observer) error {
			o.Observe(float64(runtime.NumGoroutine()))
			return nil
		}),
	); err != nil {
		return err
	}

	return nil
}
