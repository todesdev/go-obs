package goobs

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/todesdev/go-obs/internal/metrics"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"

	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/tracing"
	"github.com/todesdev/go-obs/middleware"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	FiberApp               *fiber.App
	ServiceName            string
	ServiceVersion         string
	Prometheus             bool
	EnableFiberMiddleware  bool
	EnableMetricsHandler   bool
	MetricsHandlerEndpoint string
	TracingGRPC            bool
	GRPCEndpoint           string
}

func Initialize(config *Config) error {
	validatedConfig, err := validateConfig(config)
	if err != nil {
		return err
	}

	logging.Setup(validatedConfig.ServiceName, validatedConfig.ServiceVersion)

	res, err := registerResource(validatedConfig.ServiceName, validatedConfig.ServiceVersion)
	if err != nil {
		return err
	}

	if validatedConfig.TracingGRPC {
		err := tracing.SetupOtlpGrpcTracer(validatedConfig.GRPCEndpoint, validatedConfig.ServiceName, res)

		if err != nil {
			return err
		}
	} else {
		err := tracing.SetupStdOutTracer(validatedConfig.ServiceName, res)

		if err != nil {
			return err
		}
	}

	promRegistry := &prometheus.Registry{}

	if validatedConfig.Prometheus {
		promRegistry = metrics.Setup(validatedConfig.ServiceName)
	}

	if validatedConfig.EnableFiberMiddleware {
		registerFiberMiddleware(validatedConfig.FiberApp)
	}

	if validatedConfig.EnableMetricsHandler {
		registerFiberMetricsHandler(validatedConfig.FiberApp, promRegistry, validatedConfig.MetricsHandlerEndpoint)
	}

	return nil
}

func validateConfig(cfg *Config) (*Config, error) {
	var validatedConfig Config

	validatedConfig.TracingGRPC = cfg.TracingGRPC
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

	if cfg.ServiceVersion == "" {
		validatedConfig.ServiceVersion = "1.0.0"
	} else {
		validatedConfig.ServiceVersion = cfg.ServiceVersion
	}

	if cfg.MetricsHandlerEndpoint == "" {
		validatedConfig.MetricsHandlerEndpoint = "/metrics"
	} else {
		validatedConfig.MetricsHandlerEndpoint = cfg.MetricsHandlerEndpoint
	}

	if cfg.Prometheus {
		validatedConfig.Prometheus = cfg.Prometheus
	}

	if cfg.TracingGRPC && cfg.GRPCEndpoint == "" {
		return nil, errors.New("tracing push mode is enabled but grpc endpoint is empty")
	} else {
		validatedConfig.GRPCEndpoint = cfg.GRPCEndpoint
	}

	return &validatedConfig, nil
}

func registerResource(serviceName, serviceVersion string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
		))
}

func registerFiberMiddleware(fiberApp *fiber.App) {
	fiberApp.Use(middleware.Observability())
}

func registerFiberMetricsHandler(fiberApp *fiber.App, registry *prometheus.Registry, metricsEndpoint string) {
	metricsHandler := adaptor.HTTPHandler(promhttp.HandlerFor(registry, promhttp.HandlerOpts{}))
	fiberApp.Get(metricsEndpoint, metricsHandler)
}

func NewInternalTrace(ctx context.Context, processName string) (context.Context, trace.Span) {
	return tracing.NewInternalTrace(ctx, processName)
}

func NewServerTrace(ctx context.Context, processName string) (context.Context, trace.Span) {
	return tracing.NewServerTrace(ctx, processName)
}

func NewClientTrace(ctx context.Context, processName string) (context.Context, trace.Span) {
	return tracing.NewClientTrace(ctx, processName)
}

func NewProducerTrace(ctx context.Context, processName string) (context.Context, trace.Span) {
	return tracing.NewProducerTrace(ctx, processName)
}

func NewConsumerTrace(ctx context.Context, processName string) (context.Context, trace.Span) {
	return tracing.NewConsumerTrace(ctx, processName)
}

func LoggerWithProcess(processName string) *logging.Logger {
	return logging.LoggerWithProcess(processName)
}

func TracedLoggerWithProcess(span trace.Span, processName string) *logging.Logger {
	return logging.TracedLoggerWithProcess(span, processName)
}
