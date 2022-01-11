package server

import (
	"github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/options"

	"github.com/grpc-ecosystem/go-grpc-middleware"
	recovery "github.com/grpc-ecosystem/go-grpc-middleware/recovery"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func NewServer(opts ...options.ServerOption) *grpc.Server {
	interceptors := options.NewServerOptions(opts...)
	interceptors = append(interceptors, recovery.UnaryServerInterceptor(recovery.WithRecoveryHandler(func(p interface{}) (err error) {
		return status.Errorf(codes.Internal, "panic triggered: %v", p)
	})))
	interceptors = append(interceptors, errors.ServerErrorInterceptor)
	return grpc.NewServer(grpc.UnaryInterceptor(grpc_middleware.ChainUnaryServer(interceptors...)))
}
