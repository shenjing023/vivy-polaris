package registry

import (
	"context"
	"fmt"
	"net"

	clientv3 "go.etcd.io/etcd/client/v3"
	"google.golang.org/grpc"
)

type etcdRegister struct {
	cli *clientv3.Client
	ttl int64
	key string
}

type Option func(*etcdRegister)

func WithTTL(ttl int64) Option {
	return func(r *etcdRegister) {
		r.ttl = ttl
	}
}

func NewEtcdRegister(conf clientv3.Config, serviceDesc grpc.ServiceDesc, host, port string, opts ...Option) (*etcdRegister, error) {
	cli, err := clientv3.New(conf)
	if err != nil {
		return nil, fmt.Errorf("create etcd clientv3 client failed, errmsg:%v", err)
	}

	r := &etcdRegister{
		cli: cli,
		ttl: 10,
	}
	for _, o := range opts {
		o(r)
	}

	//lease
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	resp, err := cli.Grant(ctx, r.ttl)
	if err != nil {
		return nil, fmt.Errorf("grant failed, errmsg:%v", err)
	}

	//  schema:///serviceName/ip:port ->ip:port
	serviceValue := net.JoinHostPort(host, port)
	serviceKey := getPrefix(serviceDesc) + serviceValue
	r.key = serviceKey

	//set key->value
	if _, err := cli.Put(ctx, serviceKey, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
		return nil, fmt.Errorf("put failed, errmsg:%v， key:%s, value:%s", err, serviceKey, serviceValue)
	}

	//keepalive
	kresp, err := cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return nil, fmt.Errorf("keepalive faild, errmsg:%v, lease id:%d", err, resp.ID)
	}

	go func() {
	FLOOP:
		for v := range kresp {
			if v == nil {
				fmt.Println("etcd keepalive closed")
				break FLOOP
			}
		}
	}()

	return r, nil
}

// Close 注销服务
func (r *etcdRegister) Deregister() error {
	if _, err := r.cli.Delete(context.Background(), r.key); err != nil {
		return err
	}
	return r.cli.Close()
}
