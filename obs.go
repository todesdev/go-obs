package goobs

import (
	"context"
	"errors"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/todesdev/go-obs/interceptors"
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/metrics"
	"github.com/todesdev/go-obs/internal/observer"
	"github.com/todesdev/go-obs/internal/tracing"
	"github.com/todesdev/go-obs/middleware"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
	"google.golang.org/grpc"
)

type Config struct {
	FiberApp               *fiber.App
	ServiceName            string
	ServiceVersion         string
	Region                 string
	OTLPGRPCEndpoint       string
	TracingEnabled         bool
	MetricsEnabled         bool
	MetricsHandlerEndpoint string
	MetricsHTTP            bool
	MetricsGRPC            bool
	MetricsNATS            bool
}

func Initialize(config *Config) error {
	validatedConfig, err := validateConfig(config)
	if err != nil {
		return err
	}

	logging.Setup(validatedConfig.Region, validatedConfig.ServiceName, validatedConfig.ServiceVersion)

	if validatedConfig.TracingEnabled {
		observer.SetTracingEnabled(true)
		res, err := registerResource(validatedConfig.ServiceName, validatedConfig.ServiceVersion, validatedConfig.Region)
		if err != nil {
			return err
		}

		if validatedConfig.OTLPGRPCEndpoint != "" {
			err := tracing.SetupOtlpGrpcTracer(validatedConfig.OTLPGRPCEndpoint, validatedConfig.ServiceName, res)
			if err != nil {
				return err
			}
		} else {
			err := tracing.SetupStdOutTracer(validatedConfig.ServiceName, res)

			if err != nil {
				return err
			}
		}
	}

	if validatedConfig.MetricsEnabled {
		promRegistry := &prometheus.Registry{}

		promRegistry = metrics.Setup(validatedConfig.ServiceName, validatedConfig.MetricsHTTP, validatedConfig.MetricsGRPC, validatedConfig.MetricsNATS)

		registerFiberMiddleware(validatedConfig.FiberApp, validatedConfig.TracingEnabled, validatedConfig.MetricsEnabled)
		registerFiberMetricsHandler(validatedConfig.FiberApp, promRegistry, validatedConfig.MetricsHandlerEndpoint)
	}

	return nil
}

func validateConfig(cfg *Config) (*Config, error) {
	var validatedConfig Config

	if cfg.FiberApp == nil {
		return nil, errors.New("fiber app is nil")
	} else {
		validatedConfig.FiberApp = cfg.FiberApp
	}

	validatedConfig.ServiceName = cfg.ServiceName
	validatedConfig.ServiceVersion = cfg.ServiceVersion
	validatedConfig.Region = cfg.Region
	validatedConfig.TracingEnabled = cfg.TracingEnabled
	validatedConfig.MetricsEnabled = cfg.MetricsEnabled
	validatedConfig.MetricsHTTP = cfg.MetricsHTTP
	validatedConfig.MetricsGRPC = cfg.MetricsGRPC
	validatedConfig.MetricsNATS = cfg.MetricsNATS

	validatedConfig.OTLPGRPCEndpoint = cfg.OTLPGRPCEndpoint

	if cfg.MetricsHandlerEndpoint == "" {
		validatedConfig.MetricsHandlerEndpoint = "/metrics"
	} else {
		validatedConfig.MetricsHandlerEndpoint = cfg.MetricsHandlerEndpoint
	}

	return &validatedConfig, nil
}

func registerResource(serviceName, serviceVersion, region string) (*resource.Resource, error) {
	return resource.Merge(resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String(serviceVersion),
			semconv.CloudRegion(region),
		))
}

func GRPCClientInterceptors() []grpc.DialOption {
	return []grpc.DialOption{
		grpc.WithStatsHandler(otelgrpc.NewClientHandler())}
}

func GRPCServerInterceptors() []grpc.ServerOption {
	return []grpc.ServerOption{
		grpc.StatsHandler(otelgrpc.NewServerHandler()),
		grpc.UnaryInterceptor(interceptors.UnaryServerInterceptor()),
		grpc.StreamInterceptor(interceptors.StreamServerInterceptor()),
	}
}

func registerFiberMiddleware(fiberApp *fiber.App, tracingEnabled bool, metricsEnabled bool) {
	fiberApp.Use(middleware.Observability(tracingEnabled, metricsEnabled))
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

func SimpleLogger(processName string) *logging.Logger {
	return logging.LoggerWithProcess(processName)
}

func TracedLoggerWithProcess(span trace.Span, processName string) *logging.Logger {
	return logging.TracedLoggerWithProcess(span, processName)
}
