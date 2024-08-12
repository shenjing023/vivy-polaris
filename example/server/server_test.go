package server

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"testing"
	"time"

	"github.com/shenjing023/vivy-polaris/contrib/ratelimit"
	"github.com/shenjing023/vivy-polaris/contrib/registry"
	"github.com/shenjing023/vivy-polaris/contrib/tracing"
	er "github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/example/pb"
	llog "github.com/shenjing023/vivy-polaris/log"
	vp_server "github.com/shenjing023/vivy-polaris/server"
	clientv3 "go.etcd.io/etcd/client/v3"
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
	// return nil, er.NewServiceErr(codes.Code(pb.Code_ERROR1), errors.New("test error"))
	// return nil, errors.New("test error")
	// return nil, er.NewServiceErr(codes.PermissionDenied, errors.New("test error1"))
	return nil, er.NewInternalError()
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

func TestRegistry(t *testing.T) {
	conf := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
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
	srv := vp_server.NewServer(vp_server.WithTBRL(tbp))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func TestDebug(t *testing.T) {
	llog.Init(llog.WithLevel(slog.LevelDebug))
	srv := vp_server.NewServer(vp_server.WithDebug(true))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func TestTracing(t *testing.T) {

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

	srv := vp_server.NewServer(vp_server.WithDebug(true), vp_server.WithServerTracing(tp))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Fatalf("failed to serve: %v", err)
	}
}

func TestValidate(t *testing.T) {
	srv := vp_server.NewServer(vp_server.WithServerValidator(false))
	pb.RegisterGreeterServer(srv, &test_server{})
	t.Logf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		t.Logf("failed to serve: %v", err)
	}
}
