package logging

import (
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Logger struct {
	logger *zap.Logger
}

func LoggerWithProcess(processName string) *Logger {
	mainLogger := getLogger()

	return &Logger{
		logger: mainLogger.With(zap.String("process", processName)),
	}
}

func TracedLoggerWithProcess(span trace.Span, processName string) *Logger {
	mainLogger := getLogger()
	traceID := getTraceID(span)

	return &Logger{
		logger: mainLogger.With(zap.String("process", processName), zap.String("traceID", traceID)),
	}
}

func (l *Logger) Info(msg string, fields ...zap.Field) {
	l.logger.Info(msg, fields...)
}

func (l *Logger) Warn(msg string, fields ...zap.Field) {
	l.logger.Warn(msg, fields...)
}

func (l *Logger) Error(msg string, fields ...zap.Field) {
	l.logger.Error(msg, fields...)
}

func (l *Logger) Fatal(msg string, fields ...zap.Field) {
	l.logger.Fatal(msg, fields...)
}

func getTraceID(span trace.Span) string {
	return span.SpanContext().TraceID().String()
}
