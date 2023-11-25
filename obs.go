package goobs

import (
	"context"
	"errors"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/metrics"
	"github.com/todesdev/go-obs/internal/tracing"
	"github.com/todesdev/go-obs/middleware"
	"go.opentelemetry.io/otel/trace"
)

const (
	SpanInternal = trace.SpanKindInternal
	SpanServer   = trace.SpanKindServer
	SpanClient   = trace.SpanKindClient
	SpanProducer = trace.SpanKindProducer
	SpanConsumer = trace.SpanKindConsumer
)

type Config struct {
	FiberApp               *fiber.App
	ServiceName            string
	MetricsEndpoint        string
	EnableFiberMiddleware  bool
	EnableMetricsHandler   bool
	TracingPushModeEnabled bool
	TracingGRPCEndpoint    string
}

func Initialize(config *Config) error {
	validatedConfig, err := validateConfig(config)
	if err != nil {
		return err
	}

	logging.Setup(validatedConfig.ServiceName)
	promRegistry := metrics.Setup(validatedConfig.ServiceName, validatedConfig.MetricsEndpoint)

	if validatedConfig.TracingPushModeEnabled {
		err := tracing.SetupOtlpGrpcTracer(validatedConfig.TracingGRPCEndpoint, validatedConfig.ServiceName)

		if err != nil {
			return err
		}
	} else {
		err := tracing.SetupStdOutTracer(validatedConfig.ServiceName)

		if err != nil {
			return err
		}
	}

	if validatedConfig.EnableFiberMiddleware {
		registerFiberMiddleware(validatedConfig.FiberApp)
	}

	if validatedConfig.EnableMetricsHandler {
		registerFiberMetricsHandler(validatedConfig.FiberApp, promRegistry, validatedConfig.MetricsEndpoint)
	}

	return nil
}

func validateConfig(cfg *Config) (*Config, error) {
	var validatedConfig Config

	validatedConfig.TracingPushModeEnabled = cfg.TracingPushModeEnabled
	validatedConfig.EnableFiberMiddleware = cfg.EnableFiberMiddleware
	validatedConfig.EnableMetricsHandler = cfg.EnableMetricsHandler

	if cfg.FiberApp == nil {
		return nil, errors.New("fiber app is nil")
	} else {
		validatedConfig.FiberApp = cfg.FiberApp
	}

	if cfg.ServiceName == "" {
		validatedConfig.ServiceName = "fiber_server"
	} else {
		validatedConfig.ServiceName = cfg.ServiceName
	}

	if cfg.MetricsEndpoint == "" {
		validatedConfig.MetricsEndpoint = "/metrics"
	} else {
		validatedConfig.MetricsEndpoint = cfg.MetricsEndpoint
	}

	if cfg.TracingPushModeEnabled && cfg.TracingGRPCEndpoint == "" {
		return nil, errors.New("tracing push mode is enabled but grpc endpoint is empty")
	} else {
		validatedConfig.TracingGRPCEndpoint = cfg.TracingGRPCEndpoint
	}

	// TODO: validate grpc endpoint format

	return &validatedConfig, nil
}

func registerFiberMiddleware(fiberApp *fiber.App) {
	fiberApp.Use(middleware.Observability())
}

func registerFiberMetricsHandler(fiberApp *fiber.App, registry *prometheus.Registry, metricsEndpoint string) {
	metricsHandler := adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fiberApp.Get(metricsEndpoint, metricsHandler)
}

func NewTrace(ctx context.Context, spanKind trace.SpanKind, processName string) (context.Context, trace.Span) {
	return tracing.NewTrace(ctx, spanKind, processName)
}

func LoggerWithProcess(processName string) *logging.Logger {
	return logging.LoggerWithProcess(processName)
}

func TracedLoggerWithProcess(span trace.Span, processName string) *logging.Logger {
	return logging.TracedLoggerWithProcess(span, processName)
}
