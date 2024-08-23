package registry

import (
	"context"
	"net"

	"github.com/cockroachdb/errors"
	clientv3 "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/naming/endpoints"
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

// Deprecated: Use [NewEtcdRegister] instead.
func NewEtcdRegister2(conf clientv3.Config, serviceDesc grpc.ServiceDesc, host, port string, opts ...Option) (*etcdRegister, error) {
	cli, err := clientv3.New(conf)
	if err != nil {
		return nil, errors.Errorf("create etcd clientv3 client failed: %v", err)
	}

	r := &etcdRegister{
		cli: cli,
		ttl: 10,
	}
	for _, o := range opts {
		o(r)
	}

	//lease
	ctx, cancel := context.WithTimeout(context.Background(), conf.DialTimeout)
	defer cancel()
	resp, err := cli.Grant(ctx, r.ttl)
	if err != nil {
		return nil, errors.Errorf("etcd grant failed: %v", err)
	}

	//  schema:///serviceName/ip:port ->ip:port
	serviceValue := net.JoinHostPort(host, port)
	serviceKey := getPrefix(serviceDesc) + serviceValue
	r.key = serviceKey

	//set key->value
	if _, err := cli.Put(ctx, serviceKey, serviceValue, clientv3.WithLease(resp.ID)); err != nil {
		return nil, errors.Errorf("etcd put failed, errmsg:%v， key:%s, value:%s", err, serviceKey, serviceValue)
	}

	//keepalive
	kresp, err := cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return nil, errors.Errorf("etcd keepalive faild, errmsg:%v, lease id:%d", err, resp.ID)
	}

	go func() {
	FLOOP:
		for v := range kresp {
			if v == nil {
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

func NewEtcdRegister(conf clientv3.Config, serviceDesc grpc.ServiceDesc, host, port string, opts ...Option) (*etcdRegister, error) {
	cli, err := clientv3.New(conf)
	if err != nil {
		return nil, errors.Errorf("create etcd clientv3 client failed: %v", err)
	}

	r := &etcdRegister{
		cli: cli,
		ttl: 10,
	}
	for _, o := range opts {
		o(r)
	}

	etcdManager, _ := endpoints.NewManager(cli, serviceDesc.ServiceName)

	//lease
	ctx, cancel := context.WithTimeout(context.Background(), conf.DialTimeout)
	defer cancel()
	resp, err := cli.Grant(ctx, r.ttl)
	if err != nil {
		return nil, errors.Errorf("etcd grant failed: %v", err)
	}

	//  serviceName/ip:port ->ip:port
	serviceValue := net.JoinHostPort(host, port)
	serviceKey := serviceDesc.ServiceName + "/" + serviceValue
	r.key = serviceKey

	err = etcdManager.AddEndpoint(ctx, serviceKey, endpoints.Endpoint{Addr: serviceValue}, clientv3.WithLease(resp.ID))
	if err != nil {
		return nil, errors.Errorf("etcd add endpoint failed: %v", err)
	}

	//keepalive
	kresp, err := cli.KeepAlive(context.Background(), resp.ID)
	if err != nil {
		return nil, errors.Errorf("etcd keepalive faild, errmsg:%v, lease id:%d", err, resp.ID)
	}

	go func() {
	FLOOP:
		for v := range kresp {
			if v == nil {
				break FLOOP
			}
		}
	}()

	return r, nil
}
