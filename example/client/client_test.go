package client

import (
	"context"
	"testing"
	"time"

	vp_client "github.com/shenjing023/vivy-polaris/client"
	"github.com/shenjing023/vivy-polaris/contrib/registry"
	"github.com/shenjing023/vivy-polaris/contrib/tracing"
	er "github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/example/pb"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
	addr = "localhost:50051"
	name = "world"
)

func TestClient(t *testing.T) {
	// Set up a connection to the server.
	conn, err := vp_client.NewClientConn(addr, vp_client.WithInsecure())
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
		switch s.Code() {
		case codes.Code(pb.Code_ERROR1):
			t.Log(s.Message())
		case codes.Code(pb.Code_ERROR2):
			t.Log(s.Message())
		case codes.PermissionDenied:
			t.Log(s.Message())
		case codes.Internal:
			t.Log(s.Message())
		default:
			t.Log(s.Code())
		}
		t.Fatalf("could not greet: %v", s)
	}
	t.Logf("Greeting: %s", r.GetMessage())
}

func TestDiscovery(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	}
	conn, err := vp_client.NewClientConn(registry.GetServiceTarget(pb.Greeter_ServiceDesc), vp_client.WithEtcdDiscovery(conf, pb.Greeter_ServiceDesc),
		vp_client.WithInsecure(), vp_client.WithRRLB())
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

// func TestRetry(t *testing.T) {
// 	conf := clientv3.Config{
// 		Endpoints:   []string{"10.0.0.215:2379"},
// 		DialTimeout: time.Second * 5,
// 	}
// 	retry := options.RetryPolicy{
// 		MaxAttempts:          3,
// 		MaxBackoff:           "3s",
// 		InitialBackoff:       ".1s",
// 		BackoffMultiplier:    5,
// 		RetryableStatusCodes: []string{common.CodeMap[common.CUSTOM_ERR_CODE1]},
// 	}
// 	mc := options.MethodConfig{
// 		Name: []options.MethodName{
// 			{Service: pb.Greeter_ServiceDesc.ServiceName, Method: "SayHello"},
// 		},
// 		RetryPolicy: retry,
// 	}
// 	conn, err := vp_client.NewClientConn(registry.GetServiceTarget(pb.Greeter_ServiceDesc), options.WithEtcdDiscovery(conf, pb.Greeter_ServiceDesc),
// 		options.WithInsecure(), options.WithRRLB(), options.WithRetry(mc))
// 	if err != nil {
// 		t.Fatalf("net.Connect err: %v", err)
// 	}
// 	defer conn.Close()
// 	c := pb.NewGreeterClient(conn)
// 	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
// 	defer cancel()
// 	resp, err := c.SayHello(ctx, &pb.HelloRequest{Name: name})
// 	if err != nil {
// 		s := er.Convert(err)
// 		t.Fatalf("could not greet: %v", s)
// 	}
// 	t.Logf("Greeting: %s", resp.GetMessage())
// }

func TestRateLimit(t *testing.T) {
	// Set up a connection to the server.
	conn, err := vp_client.NewClientConn(addr, vp_client.WithInsecure())
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
	tp, err := tracing.NewOTLPTracerProvider(jaegerCollectURL, "test-client")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	conn, err := vp_client.NewClientConn(addr, vp_client.WithInsecure(), vp_client.WithClientTracing(tp))
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

func TestError(t *testing.T) {
	// Set up a connection to the server.
	conn, err := vp_client.NewClientConn(addr, vp_client.WithInsecure())
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
		s, _ := status.FromError(err)
		for _, detail := range s.Details() {
			switch v := detail.(type) {
			case *pb.Error:
				t.Logf("Error: %s", v.Message)
			default:
				t.Logf("Error: %+v", v)
			}
		}
		t.Fatalf("could not greet: %+v", s)
	}
	t.Logf("Greeting: %s", r.GetMessage())
}

func TestRetry2(t *testing.T) {
	retry := vp_client.RetryPolicy{
		MaxAttempts:          3,
		MaxBackoff:           "3s",
		InitialBackoff:       ".1s",
		BackoffMultiplier:    5,
		RetryableStatusCodes: []string{pb.Code_name[int32(pb.Code_ERROR1)]},
		// RetryableStatusCodes: []string{common.CodeMap[common.CUSTOM_ERR_CODE1]},
	}
	mc := vp_client.MethodConfig{
		Name: []vp_client.MethodName{
			{Service: pb.Greeter_ServiceDesc.ServiceName, Method: "SayHello"},
		},
		RetryPolicy: retry,
	}
	conn, err := vp_client.NewClientConn(addr, vp_client.WithInsecure(), vp_client.WithRetry(mc))
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
		t.Fatalf("could not greet: %+v", s.Message())
	}
	t.Logf("Greeting: %s", resp.GetMessage())
}

func TestValidator(t *testing.T) {
	// Set up a connection to the server.
	conn, err := vp_client.NewClientConn(addr, vp_client.WithInsecure(), vp_client.WithClientValidator(true))
	if err != nil {
		t.Fatalf("did not connect: %v", err)
	}
	defer conn.Close()
	c := pb.NewGreeterClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	r, err := c.SayHello(ctx, &pb.HelloRequest{
		Name: name,
		Id:   1111,
		Id2:  222,
	})
	if err != nil {
		s := er.Convert(err)
		t.Fatalf("could not greet: %v", s)
	}
	t.Logf("Greeting: %s", r.GetMessage())
}
