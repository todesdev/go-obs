package otlp_metrics

import (
	"context"
	"github.com/todesdev/go-obs/internal/otlp_metrics/http_metrics"
	"github.com/todesdev/go-obs/internal/otlp_metrics/system_metrics"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func SetupOTLPMetricGRPCExporter(metricsGRPCEndpoint string, res *resource.Resource) error {
	ctx := context.Background()

	conn, err := grpc.Dial(metricsGRPCEndpoint,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	metricsExporter, err := otlpmetricgrpc.New(ctx, otlpmetricgrpc.WithGRPCConn(conn))
	if err != nil {
		panic(err)
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(
			metric.NewPeriodicReader(metricsExporter),
		),
	)
	otel.SetMeterProvider(meterProvider)

	err = system_metrics.Setup()
	if err != nil {
		return err
	}

	err = http_metrics.Setup()
	if err != nil {
		return err
	}

	return nil
}

func SetupPrometheusExporter(res *resource.Resource) error {
	exporter, err := prometheus.New()
	if err != nil {
		return err
	}

	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(exporter))

	otel.SetMeterProvider(meterProvider)

	err = system_metrics.Setup()
	if err != nil {
		return err
	}

	err = http_metrics.Setup()
	if err != nil {
		return err
	}

	return nil
}
