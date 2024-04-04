package goobs

import (
	"context"
	"github.com/todesdev/go-obs/internal/observer"
	"runtime"
)

func InternalObserver(ctx context.Context, process ...string) *observer.Observer {
	var p string
	if len(process) > 0 {
		p = process[0]
	} else {
		pc, _, _, _ := runtime.Caller(1)
		p = runtime.FuncForPC(pc).Name()
	}

	return observer.InternalObserver(ctx, p)
}

func ServerObserver(ctx context.Context, process ...string) *observer.Observer {
	var p string
	if len(process) > 0 {
		p = process[0]
	} else {
		pc, _, _, _ := runtime.Caller(1)
		p = runtime.FuncForPC(pc).Name()
	}

	return observer.ServerObserver(ctx, p)
}

func ClientObserver(ctx context.Context, process ...string) *observer.Observer {
	var p string
	if len(process) > 0 {
		p = process[0]
	} else {
		pc, _, _, _ := runtime.Caller(1)
		p = runtime.FuncForPC(pc).Name()
	}

	return observer.ClientObserver(ctx, p)
}

func ProducerObserver(ctx context.Context, process ...string) *observer.Observer {
	var p string
	if len(process) > 0 {
		p = process[0]
	} else {
		pc, _, _, _ := runtime.Caller(1)
		p = runtime.FuncForPC(pc).Name()
	}

	return observer.ProducerObserver(ctx, p)
}

func ConsumerObserver(ctx context.Context, process ...string) *observer.Observer {
	var p string
	if len(process) > 0 {
		p = process[0]
	} else {
		pc, _, _, _ := runtime.Caller(1)
		p = runtime.FuncForPC(pc).Name()
	}

	return observer.ConsumerObserver(ctx, p)
}
