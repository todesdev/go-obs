package goobs

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
	obs := &Observer{
		ctx: ctx,
	}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeInternal(p)
}

func ServerObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{
		ctx: ctx,
	}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeServer(p)
}

func ClientObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{
		ctx: ctx,
	}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeClient(p)
}

func ProducerObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{
		ctx: ctx,
	}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeProducer(p)
}

func ConsumerObserver(ctx context.Context, processPrefix ...string) *Observer {
	obs := &Observer{
		ctx: ctx,
	}

	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return obs.observeConsumer(p)
}

func (o *Observer) observeInternal(process string) *Observer {
	c, s := tracing.NewInternalTrace(o.ctx, process)
	l := TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeServer(process string) *Observer {
	c, s := tracing.NewServerTrace(o.ctx, process)
	l := TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeClient(process string) *Observer {
	c, s := tracing.NewClientTrace(o.ctx, process)
	l := TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeProducer(process string) *Observer {
	c, s := tracing.NewProducerTrace(o.ctx, process)
	l := TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) observeConsumer(process string) *Observer {
	c, s := tracing.NewConsumerTrace(o.ctx, process)
	l := TracedLoggerWithProcess(s, process)
	o.ctx = c
	o.span = s
	o.log = l
	return o
}

func (o *Observer) RecordInfo(msg string, fields ...zap.Field) {
	o.span.SetStatus(codes.Ok, msg)
	o.log.Info(msg, fields...)
}

func (o *Observer) RecordWarning(msg string, fields ...zap.Field) {
	o.span.SetStatus(codes.Error, msg)
	o.log.Warn(msg, fields...)
}

func (o *Observer) RecordError(msg string, err error, fields ...zap.Field) {
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

func (o *Observer) Ctx() context.Context {
	return o.ctx
}

func (o *Observer) End() {
	o.span.End()
}
