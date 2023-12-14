package goobs

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/todesdev/go-obs/internal/metrics"
	"github.com/todesdev/go-obs/internal/otlp_metrics"
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

const (
	SpanInternal = trace.SpanKindInternal
	SpanServer   = trace.SpanKindServer
	SpanClient   = trace.SpanKindClient
	SpanProducer = trace.SpanKindProducer
	SpanConsumer = trace.SpanKindConsumer
)

type Config struct {
	FiberApp              *fiber.App
	ServiceName           string
	ServiceVersion        string
	MetricsEndpoint       string
	EnableFiberMiddleware bool
	EnableMetricsHandler  bool
	MetricsPrometheus     bool
	TracingGRPC           bool
	GRPCEndpoint          string
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

	if validatedConfig.MetricsPrometheus {
		err = otlp_metrics.SetupPrometheusExporter(res)
		if err != nil {
			return err
		}
	}

	//promRegistry := metrics.Setup(validatedConfig.ServiceName, validatedConfig.MetricsEndpoint)

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

	//if validatedConfig.EnableMetricsHandler {
	//	registerOTLPPrometheusMetricsHandler(validatedConfig.FiberApp, validatedConfig.MetricsEndpoint)
	//}

	promRegistry := &prometheus.Registry{}

	if validatedConfig.EnableFiberMiddleware {
		//validatedConfig.FiberApp.Use(middleware.ObservabilityOTLP())
		promRegistry = metrics.Setup(validatedConfig.ServiceName, validatedConfig.MetricsEndpoint)
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

	if cfg.MetricsEndpoint == "" {
		validatedConfig.MetricsEndpoint = "/metrics"
	} else {
		validatedConfig.MetricsEndpoint = cfg.MetricsEndpoint
	}

	//if cfg.MetricsPrometheus && cfg.GRPCEndpoint == "" {
	//	return nil, errors.New("metrics push mode is enabled but grpc endpoint is empty")
	//} else {
	//	validatedConfig.GRPCEndpoint = cfg.GRPCEndpoint
	//}

	if cfg.MetricsPrometheus {
		validatedConfig.MetricsPrometheus = cfg.MetricsPrometheus
	}

	if cfg.TracingGRPC && cfg.GRPCEndpoint == "" {
		return nil, errors.New("tracing push mode is enabled but grpc endpoint is empty")
	} else {
		validatedConfig.GRPCEndpoint = cfg.GRPCEndpoint
	}

	// TODO: validate grpc endpoint format

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

func registerOTLPPrometheusMetricsHandler(fiberApp *fiber.App, metricsEndpoint string) {
	metricsHandler := adaptor.HTTPHandler(promhttp.Handler())
	fiberApp.Get(metricsEndpoint, metricsHandler)
}

func NewTrace(ctx context.Context, spanKind trace.SpanKind, processName string) (context.Context, trace.Span) {
	return tracing.NewTrace(ctx, spanKind, processName)
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
