package goobs

import (
	"context"
	"github.com/todesdev/go-obs/internal/observer"
)

func InternalObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	return observer.InternalObserver(ctx, processPrefix...)
}

func ServerObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	return observer.ServerObserver(ctx, processPrefix...)
}

func ClientObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	return observer.ClientObserver(ctx, processPrefix...)
}

func ProducerObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	return observer.ProducerObserver(ctx, processPrefix...)
}

func ConsumerObserver(ctx context.Context, processPrefix ...string) *observer.Observer {
	return observer.ConsumerObserver(ctx, processPrefix...)
}
