package observer

import (
	"context"
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/tracing"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

type Observer struct {
	ctx  context.Context
	span trace.Span
	log  *logging.Logger
}

var tracingEnabled = false

func SetTracingEnabled(enabled bool) {
	tracingEnabled = enabled
}

func InternalObserver(ctx context.Context, process string) *Observer {
	obs := &Observer{}

	return obs.observeInternal(ctx, process)
}

func ServerObserver(ctx context.Context, process string) *Observer {
	obs := &Observer{}

	return obs.observeServer(ctx, process)
}

func ClientObserver(ctx context.Context, process string) *Observer {
	obs := &Observer{}

	return obs.observeClient(ctx, process)
}

func ProducerObserver(ctx context.Context, process string) *Observer {
	obs := &Observer{}

	return obs.observeProducer(ctx, process)
}

func ConsumerObserver(ctx context.Context, process string) *Observer {
	obs := &Observer{}

	return obs.observeConsumer(ctx, process)
}

func (o *Observer) observeInternal(ctx context.Context, process string) *Observer {
	if tracingEnabled {
		c, s := tracing.NewInternalTrace(ctx, process)
		l := logging.TracedLoggerWithProcess(s, process)
		o.ctx = c
		o.span = s
		o.log = l
		return o
	}

	l := logging.LoggerWithProcess(process)
	o.ctx = ctx
	o.log = l
	return o
}

func (o *Observer) observeServer(ctx context.Context, process string) *Observer {
	if tracingEnabled {
		c, s := tracing.NewServerTrace(ctx, process)
		l := logging.TracedLoggerWithProcess(s, process)
		o.ctx = c
		o.span = s
		o.log = l
		return o
	}

	l := logging.LoggerWithProcess(process)
	o.ctx = ctx
	o.log = l
	return o
}

func (o *Observer) observeClient(ctx context.Context, process string) *Observer {
	if tracingEnabled {
		c, s := tracing.NewClientTrace(ctx, process)
		l := logging.TracedLoggerWithProcess(s, process)
		o.ctx = c
		o.span = s
		o.log = l
		return o
	}

	l := logging.LoggerWithProcess(process)
	o.ctx = ctx
	o.log = l
	return o
}

func (o *Observer) observeProducer(ctx context.Context, process string) *Observer {
	if tracingEnabled {
		c, s := tracing.NewProducerTrace(ctx, process)
		l := logging.TracedLoggerWithProcess(s, process)
		o.ctx = c
		o.span = s
		o.log = l
		return o
	}

	l := logging.LoggerWithProcess(process)
	o.ctx = ctx
	o.log = l
	return o
}

func (o *Observer) observeConsumer(ctx context.Context, process string) *Observer {
	if tracingEnabled {
		c, s := tracing.NewConsumerTrace(ctx, process)
		l := logging.TracedLoggerWithProcess(s, process)
		o.ctx = c
		o.span = s
		o.log = l
		return o
	}

	l := logging.LoggerWithProcess(process)
	o.ctx = ctx
	o.log = l
	return o
}

func (o *Observer) RecordInfo(msg string) {
	if tracingEnabled {
		o.span.SetStatus(codes.Ok, msg)
	}
}

func (o *Observer) RecordInfoWithLogging(msg string, fields ...zap.Field) {
	if tracingEnabled {
		o.span.SetStatus(codes.Ok, msg)
	}
	o.log.Info(msg, fields...)
}

func (o *Observer) RecordError(err error) {
	if tracingEnabled {
		o.span.RecordError(err)
	}
}

func (o *Observer) RecordErrorWithLogging(msg string, err error, fields ...zap.Field) {
	if tracingEnabled {
		o.span.RecordError(err)
	}

	fields = append(fields, zap.Error(err))
	o.log.Error(msg, fields...)
}

func (o *Observer) LogInfo(msg string, fields ...zap.Field) {
	o.log.Info(msg, fields...)
}

func (o *Observer) LogWarning(msg string, fields ...zap.Field) {
	o.log.Warn(msg, fields...)
}

func (o *Observer) LogError(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	o.log.Error(msg, fields...)
}

func (o *Observer) LogFatal(msg string, err error, fields ...zap.Field) {
	fields = append(fields, zap.Error(err))
	o.log.Fatal(msg, fields...)
}

func (o *Observer) Ctx() context.Context {
	return o.ctx
}

func (o *Observer) End() {
	if tracingEnabled {
		o.span.End()
	}
}
