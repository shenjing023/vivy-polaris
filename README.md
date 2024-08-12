# vivy-polaris
The project name has nothing to do with the actual content of this framework, just enjoyed the musical interlude &lt;&lt;Ensemble for Polaris&gt;&gt; of <<Vivy: Fluorite Eye's Song>> recently.

# TODO
- [x] 1. 添加模板，使用cli生成可执行的demo服务
- [ ] 2. 使用gapic-generator-go自动生成client，参考：[gapic-generator-go](https://github.com/googleapis/gapic-generator-go) [
gapic_generator_grpc_helloworld
](https://github.com/salrashid123/gapic_generator_grpc_helloworld)

# Useage
server
```go
import (
    vp_server "github.com/shenjing023/vivy-polaris/server"
    "log"
    "net"
    "os"
    "os/signal"
	"syscall"

    pb "xxxx"
)

func main() {
    lis, err := net.Listen("tcp", fmt.Sprintf(":%d", 8888))
	if err != nil {
		log.Fatalf("failed to listen: %+v", err)
	}
	s := vp_server.NewServer()
	pb.RegisterXXXXXServer(s, &handler.Server{})
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

```
