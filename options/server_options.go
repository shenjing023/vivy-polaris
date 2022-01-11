package options

import (
	"context"
	"time"

	"github.com/shenjing023/vivy-polaris/contrib/ratelimit"

	log "github.com/shenjing023/llog"
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

func WithDebug() ServerOption {
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
