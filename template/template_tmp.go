pakcage main

import (
    "flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"path"
	"syscall"

    log "github.com/shenjing023/llog"
	conf "github.com/shenjing023/vivy-polaris/template/config"
	vp_server "github.com/shenjing023/vivy-polaris/server"
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
	pb.RegisterGreeterServer(s, &handler.Server{})
	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf("failed to serve: %+v", err)
		}
	}()
	log.Infof("%s server start success", conf.ServerCfg.ServerName)

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	<-quit
	log.Infof("signal %d received and shutdown service", quit)
	s.GracefulStop()
}