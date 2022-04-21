package server

import (
	"context"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"reflect"
	"syscall"
	"testing"
	"time"

	"bou.ke/monkey"
	"github.com/shenjing023/vivy-polaris/contrib/ratelimit"
	"github.com/shenjing023/vivy-polaris/contrib/registry"
	"github.com/shenjing023/vivy-polaris/contrib/tracing"
	er "github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/example/common"
	"github.com/shenjing023/vivy-polaris/example/pb"
	"github.com/shenjing023/vivy-polaris/options"
	vp_server "github.com/shenjing023/vivy-polaris/server"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc/codes"
)

var (
	port = 50051
	lis  net.Listener
	err  error
	host = "localhost"
)

type test_server struct {
	pb.UnimplementedGreeterServer
}

// SayHello implements helloworld.GreeterServer
func (s *test_server) SayHello(ctx context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
	log.Printf("Received: %v", in.GetName())
	// return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
	return nil, er.NewServiceErr(common.CUSTOM_ERR_CODE1, errors.New("test error"))
}

func init() {
	lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}

func TestServer(t *testing.T) {
	srv := vp_server.NewServer()
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func TestError(t *testing.T) {
	var ts test_server
	monkey.PatchInstanceMethod(reflect.TypeOf(&ts), "SayHello", func(_ *test_server, _ context.Context, _ *pb.HelloRequest) (*pb.HelloReply, error) {
		return nil, er.NewServiceErr(common.CUSTOM_ERR_CODE1, errors.New("test error"))
	})
	srv := vp_server.NewServer()
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Logf("failed to serve: %v", err)
	}
}

func TestRegistry(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   []string{"10.0.0.215:2379"},
		DialTimeout: time.Second * 5,
	}
	r, err := registry.NewEtcdRegister(conf, pb.Greeter_ServiceDesc, host, fmt.Sprintf("%d", port))
	if err != nil {
		panic(err)
	}

	srv := vp_server.NewServer()
	pb.RegisterGreeterServer(srv, &test_server{})
	go func() {
		err = srv.Serve(lis)
		if err != nil {
			panic(err)
		}
	}()
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	for a := range c {
		switch a {
		case syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT:
			t.Log("退出")
			t.Log(r.Deregister())
			return
		default:
			return
		}
	}
}

func TestRateLimit(t *testing.T) {
	tbp := ratelimit.TBPair{Method: fmt.Sprintf("/%s/%s", pb.Greeter_ServiceDesc.ServiceName, "SayHello"), Rate: 5, Tokens: 5}
	srv := vp_server.NewServer(options.WithTBRL(tbp))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func TestDebug(t *testing.T) {
	srv := vp_server.NewServer(options.WithDebug(true))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func TestTracing(t *testing.T) {
	var ts test_server
	monkey.PatchInstanceMethod(reflect.TypeOf(&ts), "SayHello", func(_ *test_server, ctx context.Context, pr *pb.HelloRequest) (*pb.HelloReply, error) {
		return &pb.HelloReply{Message: "Hello11 " + pr.GetName()}, nil
	})

	jaegerCollectURL := "http://10.0.0.215:14268/api/traces"
	tp, err := tracing.NewJaegerTracerProvider(jaegerCollectURL, "test-server")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := tp.Shutdown(context.Background()); err != nil {
			t.Fatalf("Error shutting down tracer provider: %v", err)
		}
	}()

	srv := vp_server.NewServer(options.WithDebug(true), options.WithServerTracing(tp))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func TestError2(t *testing.T) {
	var ts test_server
	monkey.PatchInstanceMethod(reflect.TypeOf(&ts), "SayHello", func(_ *test_server, _ context.Context, in *pb.HelloRequest) (*pb.HelloReply, error) {
		// err, _ := status.New(codes.Code(pb.Code_ERROR1), "test error").WithDetails(&pb.Error{Code: pb.Code_ERROR1, Message: "custom test error"})
		log.Printf("Received: %v", in.GetName())
		return nil, er.NewServiceErr(codes.Code(pb.Code_ERROR1), errors.New("test error"))
	})
	srv := vp_server.NewServer()
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Logf("failed to serve: %v", err)
	}
}

type Number interface {
	int | float64
}
