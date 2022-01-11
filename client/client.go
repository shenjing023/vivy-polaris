package client

import (
	"context"

	"github.com/shenjing023/vivy-polaris/options"

	"google.golang.org/grpc"
)

func NewClientConnContext(ctx context.Context, target string, opts ...options.ClientOption) (*grpc.ClientConn, error) {
	// options:=[]grpc.DialOption{grpc.WithInsecure()}
	opt, err := options.NewClientOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.DialContext(ctx, target, *opt...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewClientConn(target string, opts ...options.ClientOption) (*grpc.ClientConn, error) {
	return NewClientConnContext(context.Background(), target, opts...)
}
