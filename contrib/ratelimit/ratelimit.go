package ratelimit

import (
	"context"

	"golang.org/x/time/rate"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type RateLimiter interface {
	Limit() bool
	Method() string
}

type TokenBucket struct {
	method string
	rl     *rate.Limiter
}

type TBPair struct {
	Method string // The method name
	Rate   int    // The request per second
	Tokens int    // The number of tokens in bucket
}

func (tb *TokenBucket) Limit() bool {
	return tb.rl.Allow()
}

func (tb *TokenBucket) Method() string {
	return tb.method
}

// UnaryServerInterceptor returns a new unary server interceptors that performs request rate limiting.
func UnaryServerInterceptor(limiters ...RateLimiter) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		for _, limiter := range limiters {
			if limiter.Method() == info.FullMethod {
				if !limiter.Limit() {
					return nil, status.Errorf(codes.ResourceExhausted, "method [%s] rate limit exceeded", info.FullMethod)
				}
			}
		}
		return handler(ctx, req)
	}
}

func NewTokenBucketRL(rat, tokens int, method string) RateLimiter {
	return &TokenBucket{
		method: method,
		rl:     rate.NewLimiter(rate.Limit(rat), tokens),
	}
}
