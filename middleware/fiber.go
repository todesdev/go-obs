package middleware

import (
	goobs "github.com/todesdev/go-obs"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"
	httpcollector "github.com/todesdev/go-obs/internal/metrics/http_collector"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.uber.org/zap"
)

func Observability() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if c.Path() == "/metrics" || c.Path() == "/health" || c.Path() == "/ready" {
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

		path := c.Path()
		method := c.Route().Method

		processName := "HTTP:" + method + ":" + path

		ctx := otel.GetTextMapPropagator().Extract(c.Context(), propagation.HeaderCarrier(reqHeader))

		obs := goobs.ServerObserver(ctx, processName)
		defer obs.End()

		c.SetUserContext(obs.Ctx())

		obs.LogInfo("Request received", zap.String("requestID", requestID))

		metricsCollector := httpcollector.GetHttpCollector()
		metricsCollector.IncRequestsInFlight(method)

		err := c.Next()
		if err != nil {
			elapsedTime := time.Since(startTime)

			metricsCollector.IncRequestCount(method, fiber.StatusNotFound)
			metricsCollector.ObserveResponseTime(method, fiber.StatusNotFound, elapsedTime)
			metricsCollector.DecRequestsInFlight(method)

			obs.RecordError("Request error", err)

			return c.Status(fiber.StatusNotFound).SendString("Sorry I can't find that!")
		}
		elapsedTime := time.Since(startTime)
		statusCode := c.Response().StatusCode()

		metricsCollector.IncRequestCount(method, statusCode)
		metricsCollector.ObserveResponseTime(method, statusCode, elapsedTime)
		metricsCollector.DecRequestsInFlight(method)

		obs.RecordInfo("Request completed", zap.String("requestID", requestID), zap.Int("statusCode", statusCode), zap.Duration("elapsedTime", elapsedTime))
		return nil
	}
}
