package middleware

import (
	"context"
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/observer"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	httpcollector "github.com/todesdev/go-obs/internal/metrics/http_collector"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

func Observability(tracingEnabled bool, metricsEnabled bool) fiber.Handler {
	return func(c *fiber.Ctx) error {
		if shouldSkipPath(c) {
			return c.Next()
		}

		startTime := time.Now()
		reqHeader := extractHeaders(c)
		processName := getProcessName(c)

		if tracingEnabled {
			ctx, obs := setupTracing(c, reqHeader, processName)
			defer obs.End()
			c.SetUserContext(ctx)
			obs.LogInfo("Request received")
			return processRequest(c, startTime, metricsEnabled, obs)
		}

		logger := logging.LoggerWithProcess(processName)
		logger.Info("Request received")
		return processRequest(c, startTime, metricsEnabled, logger)
	}
}

func shouldSkipPath(c *fiber.Ctx) bool {
	pathsToSkip := []string{"/metrics", "/health", "/ready"}
	for _, path := range pathsToSkip {
		if c.Path() == path {
			return true
		}
	}
	return false
}

func extractHeaders(c *fiber.Ctx) http.Header {
	reqHeader := make(http.Header)
	c.Request().Header.VisitAll(func(k, v []byte) {
		reqHeader.Add(string(k), string(v))
	})
	return reqHeader
}

func getProcessName(c *fiber.Ctx) string {
	path := c.Path()
	method := c.Route().Method
	return "HTTP:" + method + ":" + path
}

func setupTracing(c *fiber.Ctx, reqHeader http.Header, processName string) (context.Context, *observer.Observer) {
	ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(reqHeader))
	obs := observer.ServerObserver(ctx, processName)
	return obs.Ctx(), obs
}

func processRequest(c *fiber.Ctx, startTime time.Time, metricsEnabled bool, loggerOrObserver interface{}) error {
	if metricsEnabled {
		metricsCollector := httpcollector.GetHttpCollector()
		metricsCollector.IncRequestsInFlight(c.Method())
		defer metricsCollector.DecRequestsInFlight(c.Method())
	}

	err := c.Next()

	elapsedTime := time.Since(startTime)
	statusCode := c.Response().StatusCode()

	if metricsEnabled {
		metricsCollector := httpcollector.GetHttpCollector()
		metricsCollector.IncRequestCount(c.Method(), statusCode)
		metricsCollector.ObserveResponseTime(c.Method(), statusCode, elapsedTime)
	}

	switch v := loggerOrObserver.(type) {
	case *observer.Observer:
		if err != nil {
			v.RecordErrorWithLogging("Request error", err)
		} else {
			v.RecordInfoWithLogging("Request completed", zap.Int("statusCode", statusCode), zap.Duration("elapsedTime", elapsedTime))
		}
	case *zap.Logger:
		if err != nil {
			v.Error("Request error", zap.Error(err))
		} else {
			v.Info("Request completed", zap.Int("statusCode", statusCode), zap.Duration("elapsedTime", elapsedTime))
		}
	}

	if err != nil {
		return c.Status(fiber.StatusNotFound).SendString("Sorry I can't find that!")
	}

	return nil
}
