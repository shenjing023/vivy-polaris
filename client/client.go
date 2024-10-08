package client

import (
	"context"

	"github.com/shenjing023/vivy-polaris/options"

	"google.golang.org/grpc"
)

func NewClientConnContext(ctx context.Context, target string, opts ...options.Option[clientOptions]) (*grpc.ClientConn, error) {
	// options:=[]grpc.DialOption{grpc.WithInsecure()}
	opt, err := NewClientOptions(opts...)
	if err != nil {
		return nil, err
	}
	conn, err := grpc.NewClient(target, *opt...)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

func NewClientConn(target string, opts ...options.Option[clientOptions]) (*grpc.ClientConn, error) {
	return NewClientConnContext(context.Background(), target, opts...)
}
