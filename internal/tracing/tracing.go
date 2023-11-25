package tracing

import (
	"context"

	"github.com/todesdev/go-obs/internal/logging"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

const (
	SpanInternal = trace.SpanKindInternal
	SpanServer   = trace.SpanKindServer
	SpanClient   = trace.SpanKindClient
	SpanProducer = trace.SpanKindProducer
	SpanConsumer = trace.SpanKindConsumer
)

var service string

func SetupOtlpGrpcTracer(tracingGPRCEndpoint, serviceName string) error {
	logger := logging.LoggerWithProcess("TracingSetup")
	logger.Info("Setting up OLTP GRPC tracing")

	ctx := context.Background()
	conn, err := connectToOTLPCollector(ctx, tracingGPRCEndpoint)
	if err != nil {
		logger.Fatal("Failed to connect to OTLP collector", zap.Error(err))
		return err
	}

	tp, err := configureOtlpGrpcTraceProvider(ctx, conn, serviceName)
	if err != nil {
		logger.Fatal("Failed to configure trace provider", zap.Error(err))
		return err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	service = serviceName

	logger.Info("Tracing setup complete")
	return nil
}

func SetupStdOutTracer(serviceName string) error {
	logger := logging.LoggerWithProcess("TracingSetup")
	logger.Info("Setting up STDOUT tracing")

	tp, err := configureStdOutTraceProvider(serviceName)
	if err != nil {
		logger.Fatal("Failed to configure trace provider", zap.Error(err))
		return err
	}

	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))

	service = serviceName

	logger.Info("Tracing setup complete")
	return nil
}

func NewTrace(ctx context.Context, spanKind trace.SpanKind, processName string) (context.Context, trace.Span) {
	return otel.Tracer(service).Start(ctx, processName, trace.WithSpanKind(spanKind))
}

func connectToOTLPCollector(ctx context.Context, tracingGRPCEndpoint string) (*grpc.ClientConn, error) {
	return grpc.DialContext(ctx, tracingGRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
}

func configureOtlpGrpcTraceProvider(ctx context.Context, conn *grpc.ClientConn, serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithGRPCConn(conn))
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			)),
	), nil
}

func configureStdOutTraceProvider(serviceName string) (*sdktrace.TracerProvider, error) {
	exporter, err := stdouttrace.New()
	if err != nil {
		return nil, err
	}

	return sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(serviceName),
			)),
	), nil
}
