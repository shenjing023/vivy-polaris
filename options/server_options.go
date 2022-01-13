package options

import (
	"context"
	"time"

	"github.com/shenjing023/vivy-polaris/contrib/ratelimit"

	log "github.com/shenjing023/llog"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc"
)

// WithTBRL TokenBucketRateLimiter
func WithTBRL(pairs ...ratelimit.TBPair) ServerOption {
	var limiters []ratelimit.RateLimiter
	for _, p := range pairs {
		limiters = append(limiters, ratelimit.NewTokenBucketRL(p.Rate, p.Tokens, p.Method))
	}
	return newFuncServerOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, ratelimit.UnaryServerInterceptor(limiters...))
	})
}

func WithDebug(flag bool) ServerOption {
	if !flag {
		return newFuncServerOption(func(so *serverOptions) {})
	}
	log.SetConsoleLogger(
		log.WithLevel(log.DebugLevel),
	)
	return newFuncServerOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, func(ctx context.Context, req interface{},
			info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			startTime := time.Now()
			resp, err := handler(ctx, req)
			log.Debugf("method [%s] cost %s", info.FullMethod, time.Since(startTime))
			return resp, err
		})
	})
}

func NewServerOptions(opts ...ServerOption) []grpc.UnaryServerInterceptor {
	sopt := &serverOptions{
		interceptors: make([]grpc.UnaryServerInterceptor, 0),
	}
	for _, opt := range opts {
		opt.apply(sopt)
	}
	return sopt.interceptors
}

func WithServerTracing(tp *sdktrace.TracerProvider) ServerOption {
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return newFuncServerOption(func(so *serverOptions) {
		so.interceptors = append(so.interceptors, otelgrpc.UnaryServerInterceptor())
	})
}
