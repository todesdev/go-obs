package interceptors

import (
	"context"
	"github.com/todesdev/go-obs/internal/metrics/grpc_collector"
	"google.golang.org/grpc"
	"google.golang.org/grpc/status"
	"time"
)

func UnaryServerInterceptor() grpc.UnaryServerInterceptor {
	return func (ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		start := time.Now()
		h, err := handler(ctx, req)
		collector := grpc_collector.GetGrpcCollector()
		method := info.FullMethod
		collector.IncRequestsInFlight(method)
		defer collector.DecRequestsInFlight(method)
		collector.IncRequestCount(method, status.Code(err).String())
		collector.ObserveResponseTime(method, status.Code(err).String(), time.Since(start))
		return h, err
	}
}

func StreamServerInterceptor() grpc.StreamServerInterceptor {
	return func (srv interface{}, stream grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		start := time.Now()
		err := handler(srv, stream)
		collector := grpc_collector.GetGrpcCollector()
		method := info.FullMethod
		collector.IncRequestsInFlight(method)
		defer collector.DecRequestsInFlight(method)
		collector.IncRequestCount(method, status.Code(err).String())
		collector.ObserveResponseTime(method, status.Code(err).String(), time.Since(start))
		return err
	}
}