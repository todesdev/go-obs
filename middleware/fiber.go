package middleware

import (
	"github.com/todesdev/go-obs/internal/otlp_metrics/http_metrics"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	"github.com/todesdev/go-obs/internal/logging"
	httpcollector "github.com/todesdev/go-obs/internal/metrics/http_collector"
	"github.com/todesdev/go-obs/internal/tracing"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

func Observability() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// TODO: Figure out how to store and retrieve these endpoints from a config file
		if c.Route().Path == "/metrics" || c.Route().Path == "/health" || c.Route().Path == "/ready" {
			return c.Next()
		}

		startTime := time.Now()

		requestID := uuid.New().String()[:7]
		c.Request().Header.Set("X-Request-ID", requestID)
		c.Response().Header.Set("X-Request-ID", requestID)

		reqHeader := make(http.Header)
		c.Request().Header.VisitAll(func(k, v []byte) {
			reqHeader.Add(string(k), string(v))
		})

		path := c.Route().Path
		method := c.Route().Method

		processName := "HTTP:" + method + ":" + path

		ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(reqHeader))
		ctx, span := tracing.NewTrace(ctx, tracing.SpanServer, processName)
		defer span.End()

		c.SetUserContext(ctx)

		logger := logging.TracedLoggerWithProcess(span, processName)
		logger.Info("Request received", zap.String("requestID", requestID))

		metricsCollector := httpcollector.GetHttpCollector()
		metricsCollector.IncRequestsInFlight(method)

		err := c.Next()
		if err != nil {
			elapsedTime := time.Since(startTime)

			metricsCollector.IncRequestCount(method, fiber.StatusNotFound)
			metricsCollector.ObserveResponseTime(method, fiber.StatusNotFound, elapsedTime)
			metricsCollector.DecRequestsInFlight(method)

			logger.Error("Request error", zap.Error(err))
			span.RecordError(err)

			return c.Status(fiber.StatusNotFound).SendString("Sorry I can't find that!")
		}
		elapsedTime := time.Since(startTime)
		statusCode := c.Response().StatusCode()

		metricsCollector.IncRequestCount(method, statusCode)
		metricsCollector.ObserveResponseTime(method, statusCode, elapsedTime)
		metricsCollector.DecRequestsInFlight(method)

		logger.Info("Request completed", zap.String("requestID", requestID), zap.Int("statusCode", statusCode), zap.Duration("elapsedTime", elapsedTime))

		return nil
	}
}

func ObservabilityOTLP() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/health" || c.Path() == "/ready" {
			return c.Next()
		}

		startTime := time.Now()

		requestID := uuid.New().String()
		c.Request().Header.Set("X-Request-ID", requestID)
		c.Response().Header.Set("X-Request-ID", requestID)

		reqHeader := make(http.Header)
		c.Request().Header.VisitAll(func(k, v []byte) {
			reqHeader.Add(string(k), string(v))
		})

		path := c.Path()
		method := c.Method()

		processName := "HTTP:" + method + ":" + path

		ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(reqHeader))
		ctx, span := tracing.NewTrace(ctx, tracing.SpanServer, processName)
		defer span.End()

		c.SetUserContext(ctx)

		logger := logging.TracedLoggerWithProcess(span, processName)
		logger.Info("Request received", zap.String("requestID", requestID))

		http_metrics.IncreaseRequestsInFlight(ctx, method)

		err := c.Next()
		if err != nil {
			elapsedTime := time.Since(startTime)
			statusCode := c.Response().StatusCode()

			http_metrics.RecordRequest(ctx, method, statusCode)
			http_metrics.RecordResponseTime(ctx, method, statusCode, elapsedTime)
			http_metrics.DecreaseRequestsInFlight(ctx, method)

			logger.Error("Request error", zap.Error(err))
			span.RecordError(err)

			return c.Status(fiber.StatusNotFound).SendString("Sorry I can't find that!")
		}

		elapsedTime := time.Since(startTime)
		statusCode := c.Response().StatusCode()

		http_metrics.RecordRequest(ctx, method, statusCode)
		http_metrics.RecordResponseTime(ctx, method, statusCode, elapsedTime)
		http_metrics.DecreaseRequestsInFlight(ctx, method)

		logger.Info("Request completed", zap.String("requestID", requestID), zap.Int("statusCode", statusCode), zap.Duration("elapsedTime", elapsedTime))
		return nil
	}
}
