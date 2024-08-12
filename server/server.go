package server

import (
	"context"
	"time"

	"github.com/shenjing023/vivy-polaris/contrib/ratelimit"
	"github.com/shenjing023/vivy-polaris/contrib/validator"
	"github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"

	"log/slog"

	grpc_middleware "github.com/grpc-ecosystem/go-grpc-middleware"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewServer(opts ...options.Option[serverOptions]) *grpc.Server {
	interceptors := NewServerOptions(opts...)
	interceptors = append(interceptors, recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		return status.Errorf(codes.Internal, "panic triggered: %v", p)
	})))
	interceptors = append(interceptors, errors.ServerErrorInterceptor)
	return grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))
}

type serverOptions struct {
	interceptors []grpc.UnaryServerInterceptor
}

// WithTBRL TokenBucketRateLimiter
func WithTBRL(pairs ...ratelimit.TBPair) options.Option[serverOptions] {
	var limiters []ratelimit.RateLimiter
	for _, p := range pairs {
		limiters = append(limiters, ratelimit.NewTokenBucketRL(p.Rate, p.Tokens, p.Method))
	}
	return options.NewFuncOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, ratelimit.UnaryServerInterceptor(limiters...))
	})
}

func WithDebug(flag bool) options.Option[serverOptions] {
	if !flag {
		return options.NewFuncOption(func(so *serverOptions) {})
	}
	slog.Info("debug mode enabled")
	return options.NewFuncOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, func(ctx context.Context, req interface{},
			info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			slog.Debug("debug", "method", info.FullMethod, "req", req)
			startTime := time.Now()
			resp, err := handler(ctx, req)
			slog.Debug("debug", "method", info.FullMethod, "cost", time.Since(startTime))
			return resp, err
		})
	})
}

func NewServerOptions(opts ...options.Option[serverOptions]) []grpc.UnaryServerInterceptor {
	sopt := &serverOptions{
		interceptors: make([]grpc.UnaryServerInterceptor, 0),
	}
	for _, opt := range opts {
		opt.Apply(sopt)
	}
	return sopt.interceptors
}

func WithServerTracing(tp *sdktrace.TracerProvider) options.Option[serverOptions] {
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return options.NewFuncOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, otelgrpc.UnaryServerInterceptor())
	})
}

// WithServerValidator validate fields,
// all==true return all fields error, otherwise return first error
func WithServerValidator(all bool) options.Option[serverOptions] {
	return options.NewFuncOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, validator.UnaryServerInterceptor(all))
	})
}
