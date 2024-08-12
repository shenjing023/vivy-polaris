package options

import (
	"google.golang.org/grpc"
)

type clientOptions struct {
	opts          []grpc.DialOption
	serviceConfig ServiceConfig
}

type ClientOption interface {
	apply(*clientOptions)
}

type funcClientOption struct {
	f func(*clientOptions)
}

func (fdo *funcClientOption) apply(do *clientOptions) {
	fdo.f(do)
}

func newFuncClientOption(f func(*clientOptions)) *funcClientOption {
	return &funcClientOption{
		f: f,
	}
}

type serverOptions struct {
	interceptors []grpc.UnaryServerInterceptor
}

type ServerOption interface {
	apply(*serverOptions)
}

type funcServerOption struct {
	f func(*serverOptions)
}

func (fdo *funcServerOption) apply(do *serverOptions) {
	fdo.f(do)
}

func newFuncServerOption(f func(*serverOptions)) *funcServerOption {
	return &funcServerOption{
		f: f,
	}
}

type Option[T any] interface {
	Apply(*T)
}

type FuncOption[T any] struct {
	f func(*T)
}

func (fo *FuncOption[T]) Apply(o *T) {
	if fo.f != nil {
		fo.f(o)
	}
}

func NewFuncOption[T any](f func(*T)) *FuncOption[T] {
	return &FuncOption[T]{
		f: f,
	}
}
