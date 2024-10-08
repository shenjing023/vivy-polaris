package client

import (
	"encoding/json"

	"github.com/shenjing023/vivy-polaris/contrib/registry"
	"github.com/shenjing023/vivy-polaris/options"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/shenjing023/vivy-polaris/contrib/validator"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
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
		RetryableStatusCodes: 服务端返回什么错误码才重试，这里错误码只能是 gRPC 错误码，不支持自定义错误码。
*/

type clientOptions struct {
	opts          []grpc.DialOption
	serviceConfig ServiceConfig
}

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

func NewClientOptions(opts ...options.Option[clientOptions]) (*[]grpc.DialOption, error) {
	copt := &clientOptions{
		opts: make([]grpc.DialOption, 0),
	}
	for _, opt := range opts {
		opt.Apply(copt)
	}
	sc, err := json.Marshal(copt.serviceConfig)
	if err != nil {
		return nil, err
	}
	copt.opts = append(copt.opts, grpc.WithDefaultServiceConfig(string(sc)))
	// var kacp = keepalive.ClientParameters{
	// 	Time:                10 * time.Second, // send pings every 10 seconds if there is no activity
	// 	Timeout:             time.Second,      // wait 1 second for ping ack before considering the connection dead
	// 	PermitWithoutStream: true,             // send pings even without active streams
	// }
	// copt.opts = append(copt.opts, grpc.WithKeepaliveParams(kacp))
	return &copt.opts, nil
}

func WithInsecure() options.Option[clientOptions] {
	return options.NewFuncOption(func(o *clientOptions) {
		o.opts = append(o.opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	})
}

func WithRetry(mc ...MethodConfig) options.Option[clientOptions] {
	return options.NewFuncOption(func(o *clientOptions) {
		o.serviceConfig.Methodconfig = mc
	})
}

// round_robin load balancing policy
func WithRRLB() options.Option[clientOptions] {
	return options.NewFuncOption(func(o *clientOptions) {
		o.serviceConfig.LoadBalancingPolicy = "round_robin"
	})
}

func WithEtcdDiscovery(conf clientv3.Config, serviceDesc grpc.ServiceDesc) options.Option[clientOptions] {
	return options.NewFuncOption(func(o *clientOptions) {
		r, err := registry.NewEtcdResolver(conf, serviceDesc)
		if err != nil {
			panic(err)
		}
		o.opts = append(o.opts, grpc.WithResolvers(r))
		// resolver.Register(r)
	})
}

func WithClientTracing(tp *sdktrace.TracerProvider) options.Option[clientOptions] {
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{}))
	return options.NewFuncOption(func(o *clientOptions) {
		o.opts = append(o.opts, grpc.WithUnaryInterceptor(otelgrpc.UnaryClientInterceptor()))
	})
}

// WithClientValidator validate fields,
// all==true return all fields error, otherwise return first error
func WithClientValidator(all bool) options.Option[clientOptions] {
	return options.NewFuncOption(func(so *clientOptions) {
		so.opts = append(so.opts, grpc.WithUnaryInterceptor(validator.UnaryClientInterceptor(all)))
	})
}
