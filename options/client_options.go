package options

import (
	"encoding/json"

	"github.com/shenjing023/vivy-polaris/contrib/registry"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/resolver"
)

/*	methodConfig配置信息
	name 指定下面的配置信息作用的 RPC 服务或方法
		service: 通过服务名匹配，语法为<package>.<service> package就是proto文件中指定的package，service也是proto文件中指定的 Service Name。
		method: 匹配具体某个方法，proto文件中定义的方法名。

	retryPolicy，重试策略
		MaxAttempts: 最大尝试次数
		InitialBackoff: 默认退避时间
		MaxBackoff: 最大退避时间
		BackoffMultiplier: 退避时间增加倍率
		RetryableStatusCodes: 服务端返回什么错误码才重试
*/

type MethodName struct {
	Service string `json:"service"`
	Method  string `json:"method"`
}

type MethodConfig struct {
	Name        []MethodName `json:"name"`
	RetryPolicy RetryPolicy  `json:"retryPolicy"`
}

type RetryPolicy struct {
	MaxAttempts          int      `json:"MaxAttempts"`
	MaxBackoff           string   `json:"MaxBackoff"`
	InitialBackoff       string   `json:"InitialBackoff"`
	BackoffMultiplier    int      `json:"BackoffMultiplier"`
	RetryableStatusCodes []string `json:"RetryableStatusCodes"`
}

type ServiceConfig struct {
	Methodconfig        []MethodConfig `json:"methodConfig,omitempty"`
	LoadBalancingPolicy string         `json:"loadBalancingPolicy,omitempty"`
}

func NewClientOptions(opts ...ClientOption) (*[]grpc.DialOption, error) {
	copt := &clientOptions{
		opts: make([]grpc.DialOption, 0),
	}
	for _, opt := range opts {
		opt.apply(copt)
	}
	sc, err := json.Marshal(copt.serviceConfig)
	if err != nil {
		return nil, err
	}
	copt.opts = append(copt.opts, grpc.WithDefaultServiceConfig(string(sc)))
	return &copt.opts, nil
}

func WithInsecure() ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.opts = append(o.opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	})
}

func WithRetry(mc ...MethodConfig) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.serviceConfig.Methodconfig = mc
	})
}

// round_robin load balancing policy
func WithRRLB() ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		o.serviceConfig.LoadBalancingPolicy = "round_robin"
	})
}

func WithEtcdDiscovery(conf clientv3.Config, serviceDesc grpc.ServiceDesc) ClientOption {
	return newFuncClientOption(func(o *clientOptions) {
		r, err := registry.NewEtcdResolver(conf, serviceDesc)
		if err != nil {
			panic(err)
		}
		resolver.Register(r)
	})
}

func WithClientTracing(tp *sdktrace.TracerProvider) ClientOption {
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return newFuncClientOption(func(o *clientOptions) {
		o.opts = append(o.opts, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	})
}
