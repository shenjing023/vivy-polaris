/*
	https://github.com/envoyproxy/protoc-gen-validate
*/

package validator

import (
	"context"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type validator interface {
	Validate() error
	ValidateAll() error
}

func UnaryServerInterceptor(all bool) grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		switch v := req.(type) {
		case validator:
			if all {
				if err := v.ValidateAll(); err != nil {
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			} else {
				if err := v.Validate(); err != nil {
					return nil, status.Error(codes.InvalidArgument, err.Error())
				}
			}
		}
		return handler(ctx, req)
	}
}

func UnaryClientInterceptor(all bool) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		switch v := req.(type) {
		case validator:
			if all {
				if err := v.ValidateAll(); err != nil {
					return status.Error(codes.InvalidArgument, err.Error())
				}
			} else {
				if err := v.Validate(); err != nil {
					return status.Error(codes.InvalidArgument, err.Error())
				}
			}
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
