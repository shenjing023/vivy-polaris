package main

import (
    "flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"syscall"

    "log"
	conf "{{.PkgName}}/config"
	vp_server "github.com/shenjing023/vivy-polaris/server"
	handler "{{.PkgName}}/internal"
	pb "{{.PkgName}}/{{.GRPCPath}}"
)

var (
	confPath *string
)

func init() {
	pwd, _ := os.Getwd()
	confPath = flag.String("p", path.Join(pwd, "/config/config.yaml"), "config file path")
}

func main() {
	flag.Parse()
	conf.Init(*confPath)
	runServer()
}

func runServer() {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", conf.ServerCfg.Port))
	if err != nil {
		log.Fatalf("failed to listen: %+v", err)
	}
	s := vp_server.NewServer()
	pb.Register{{.ServerName}}Server(s, &handler.Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %+v", err)
		}
	}()
	log.Printf("%s server start success, port: %d", conf.ServerCfg.ServerName, conf.ServerCfg.Port)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Printf("signal %d received and shutdown service", quit)
	s.GracefulStop()
}