package client

import (
	"context"
	"testing"
	"time"

	vp_client "github.com/shenjing023/vivy-polaris/client"
	"github.com/shenjing023/vivy-polaris/contrib/registry"
	"github.com/shenjing023/vivy-polaris/contrib/tracing"
	er "github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/example/common"
	"github.com/shenjing023/vivy-polaris/example/pb"
	"github.com/shenjing023/vivy-polaris/options"
	clientv3 "go.etcd.io/etcd/client/v3"
)

var (
	addr = "localhost:50051"
	name = "world"
)

func TestClient(t *testing.T) {
	// Set up a connection to the server.
	conn, err := vp_client.NewClientConn(addr, options.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		s := er.Convert(err)
		t.Fatalf("could not greet: %v", s)
	}
	t.Logf("Greeting: %s", r.GetMessage())
}

func TestDiscovery(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   []string{"10.0.0.215:2379"},
		DialTimeout: time.Second * 5,
	}
	conn, err := vp_client.NewClientConn(registry.GetServiceTarget(pb.Greeter_ServiceDesc), options.WithEtcdDiscovery(conf, pb.Greeter_ServiceDesc),
		options.WithInsecure(), options.WithRRLB())
	if err != nil {
		t.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		s := er.Convert(err)
		t.Fatalf("could not greet: %v", s)
	}
	t.Logf("Greeting: %s", resp.GetMessage())
}

func TestRetry(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   []string{"10.0.0.215:2379"},
		DialTimeout: time.Second * 5,
	}
	retry := options.RetryPolicy{
		MaxAttempts:          3,
		MaxBackoff:           "3s",
		InitialBackoff:       ".1s",
		BackoffMultiplier:    5,
		RetryableStatusCodes: []string{common.CodeMap[common.CUSTOM_ERR_CODE1]},
	}
	mc := options.MethodConfig{
		Name: []options.MethodName{
			{Service: pb.Greeter_ServiceDesc.ServiceName, Method: "SayHello"},
		},
		RetryPolicy: retry,
	}
	conn, err := vp_client.NewClientConn(registry.GetServiceTarget(pb.Greeter_ServiceDesc), options.WithEtcdDiscovery(conf, pb.Greeter_ServiceDesc),
		options.WithInsecure(), options.WithRRLB(), options.WithRetry(mc))
	if err != nil {
		t.Fatalf("net.Connect err: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	resp, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		s := er.Convert(err)
		t.Fatalf("could not greet: %v", s)
	}
	t.Logf("Greeting: %s", resp.GetMessage())
}

func TestRateLimit(t *testing.T) {
	// Set up a connection to the server.
	conn, err := vp_client.NewClientConn(addr, options.WithInsecure())
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	for i := 0; i < 10; i++ {
		r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
		if err != nil {
			s := er.Convert(err)
			t.Logf("could not greet: %v", s)
		}
		t.Logf("Greeting: %s", r.GetMessage())
	}
	time.Sleep(time.Second * 1)
	r, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		s := er.Convert(err)
		t.Logf("could not greet: %v", s)
	}
	t.Logf("Greeting: %s", r.GetMessage())
}

func TestTracing(t *testing.T) {
	ctx := context.Background()

	jaegerCollectURL := "http://10.0.0.215:14268/api/traces"
	tp, err := tracing.NewJaegerTracerProvider(jaegerCollectURL, "test-client")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	conn, err := vp_client.NewClientConn(addr, options.WithInsecure(), options.WithClientTracing(tp))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	resp, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("reply:%s", resp.Message)
}
