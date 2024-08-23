package main

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	vp_client "github.com/shenjing023/vivy-polaris/client"
	"github.com/shenjing023/vivy-polaris/contrib/registry"
	er "github.com/shenjing023/vivy-polaris/errors"
	"github.com/shenjing023/vivy-polaris/example/pb"
	vp_server "github.com/shenjing023/vivy-polaris/server"
	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
)

var (
	port = 50052
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
	return &pb.HelloReply{Message: "Hello " + in.GetName()}, nil
	// return nil, er.NewServiceErr(codes.Code(pb.Code_ERROR1), errors.New("test error"))
	// return nil, errors.New("test error")
	// return nil, er.NewServiceErr(codes.PermissionDenied, errors.New("test error1"))
	// return nil, er.NewInternalError()
}

func init() {
	lis, err = net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
}

func TestServer() {
	srv := vp_server.NewServer()
	pb.RegisterGreeterServer(srv, &test_server{})
	log.Printf("server listening at %v", lis.Addr())
	if err := srv.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}

func TestRegistry() {
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
		log.Println("server start")
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
			log.Println("退出")
			log.Println(r.Deregister())
			return
		default:
			return
		}
	}
}

var ClientConn *grpc.ClientConn

func main() {
	go TestRegistry()
	httpServer()
}

func InitClient() {
	// endpoint := "etcd://127.0.0.1:2379/test"
	// a := strings.TrimPrefix(endpoint, "/")
	// log.Printf("a: %s\n", a)
	conf := clientv3.Config{
		Endpoints:   []string{"127.0.0.1:2379"},
		DialTimeout: time.Second * 5,
	}

	ClientConn, err = vp_client.NewClientConn(registry.GetServiceTarget(pb.Greeter_ServiceDesc),
		vp_client.WithEtcdDiscovery(conf, pb.Greeter_ServiceDesc),
		vp_client.WithInsecure(), vp_client.WithRRLB())
	if err != nil {
		log.Fatalf("net.Connect err: %v", err)
	}
}

func httpServer() {
	InitClient()
	engine := gin.Default()
	engine.GET("/ping", func(c *gin.Context) {
		conn := pb.NewGreeterClient(ClientConn)

		// Contact the server and print out its response.
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
		defer cancel()
		r, err := conn.SayHello(ctx, &pb.HelloRequest{Name: "name"})
		if err != nil {
			s := er.Convert(err)
			switch s.Code() {
			case codes.Code(pb.Code_ERROR1):
				log.Println(s.Message())
			case codes.Code(pb.Code_ERROR2):
				log.Println(s.Message())
			case codes.PermissionDenied:
				log.Println(s.Message())
			case codes.Internal:
				log.Println(s.Message())
			default:
				log.Println(s.Code())
			}
			log.Printf("could not greet: %v", s)
		}
		log.Printf("Greeting: %s\n", r.GetMessage())
		c.JSON(200, gin.H{
			"message": r.GetMessage(),
		})
	})
	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", 8099),
		Handler: engine,
	}
	go func() {
		log.Println("http server start")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("run service fatal: %v", err)
		}
	}()
	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Printf("signal %d received and shutdown service", quit)
	srv.Shutdown(context.Background())
}
