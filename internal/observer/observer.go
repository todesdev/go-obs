package observer

import (
	"context"
	"github.com/todesdev/go-obs/internal/logging"
	"github.com/todesdev/go-obs/internal/tracing"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"runtime"
)

type Observer struct {
	ctx  context.Context
	span trace.Span
	log  *logging.Logger
}

func InternalObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeInternal(ctx, p)
}

func ServerObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeServer(ctx, p)
}

func ClientObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeClient(ctx, p)
}

func ProducerObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeProducer(ctx, p)
}

func ConsumerObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeConsumer(ctx, p)
}

func (o *Observer) observeInternal(ctx context.Context, process string) *Observer {
	c, s := tracing.NewInternalTrace(ctx, process)
	l := logging.TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeServer(ctx context.Context, process string) *Observer {
	c, s := tracing.NewServerTrace(ctx, process)
	l := logging.TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeClient(ctx context.Context, process string) *Observer {
	c, s := tracing.NewClientTrace(ctx, process)
	l := logging.TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeProducer(ctx context.Context, process string) *Observer {
	c, s := tracing.NewProducerTrace(ctx, process)
	l := logging.TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeConsumer(ctx context.Context, process string) *Observer {
	c, s := tracing.NewConsumerTrace(ctx, process)
	l := logging.TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) RecordInfo(msg string) {
	o.span.SetStatus(codes.Ok, msg)
}

func (o *Observer) RecordInfoWithLogging(msg string, fields ...zap.Field) {
	o.span.SetStatus(codes.Ok, msg)
	o.log.Info(msg, fields...)
}

func (o *Observer) RecordError(err error) {
	o.span.RecordError(err)
}

func (o *Observer) RecordErrorWithLogging(msg string, err error, fields ...zap.Field) {
	o.span.RecordError(err)
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
	o.span.End()
}
