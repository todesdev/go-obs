package http_metrics

import (
	"context"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/metric"
	"time"
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
	meter            = otel.Meter(HttpSubsystem)
	requestCounter   metric.Int64Counter
	responseTime     metric.Float64Histogram
	requestsInFlight metric.Int64UpDownCounter
)

func Setup() error {
	err := setupRequestCounter()
	if err != nil {
		return err
	}

	err = setupResponseTime()
	if err != nil {
		return err
	}

	err = setupRequestsInFlight()
	if err != nil {
		return err
	}

	return nil
}

func setupRequestCounter() error {
	var err error
	requestCounter, err = meter.Int64Counter(
		HttpRequestsTotal,
		metric.WithDescription(HttpRequestsHelp),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	return nil
}

func RecordRequest(ctx context.Context, method string, statusCode int) {
	labels := []attribute.KeyValue{
		attribute.String(HttpMethodLabel, method),
		attribute.String(HttpStatusCodeLabel, string(rune(statusCode))),
	}

	requestCounter.Add(ctx, 1,
		metric.WithAttributes(labels...),
	)
}

func RecordResponseTime(ctx context.Context, method string, statusCode int, duration time.Duration) {
	labels := []attribute.KeyValue{
		attribute.String(HttpMethodLabel, method),
		attribute.String(HttpStatusCodeLabel, string(rune(statusCode))),
	}

	responseTime.Record(ctx, float64(duration)/float64(time.Second),
		metric.WithAttributes(labels...),
	)
}

func IncreaseRequestsInFlight(ctx context.Context, method string) {
	labels := []attribute.KeyValue{
		attribute.String(HttpMethodLabel, method),
	}

	requestsInFlight.Add(ctx, 1,
		metric.WithAttributes(labels...),
	)
}

func DecreaseRequestsInFlight(ctx context.Context, method string) {
	labels := []attribute.KeyValue{
		attribute.String(HttpMethodLabel, method),
	}

	requestsInFlight.Add(ctx, -1,
		metric.WithAttributes(labels...),
	)
}

func setupResponseTime() error {
	var err error
	responseTime, err = meter.Float64Histogram(
		HttpRequestDurationSeconds,
		metric.WithDescription(HttpRequestsDurationSecondsHelp),
		metric.WithUnit("s"),
	)
	if err != nil {
		return err
	}

	return nil
}

func setupRequestsInFlight() error {
	var err error
	requestsInFlight, err = meter.Int64UpDownCounter(
		HttpRequestsInProgressTotal,
		metric.WithDescription(HttpRequestsInProgressHelp),
		metric.WithUnit("{call}"),
	)
	if err != nil {
		return err
	}

	return nil
}
