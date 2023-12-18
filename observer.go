package goobs

import (
	"context"
	"github.com/todesdev/go-obs/internal/observer"
	"runtime"
)

func InternalObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return observer.InternalObserver(ctx, p)
}

func ServerObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return observer.ServerObserver(ctx, p)
}

func ClientObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return observer.ClientObserver(ctx, p)
}

func ProducerObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return observer.ProducerObserver(ctx, p)
}

func ConsumerObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	pc, _, _, _ := runtime.Caller(1)
	p := runtime.FuncForPC(pc).Name()

	if len(processPrefix) > 0 {
		p = processPrefix[0] + ":" + p
	}

	return observer.ConsumerObserver(ctx, p)
}
