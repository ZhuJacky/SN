// Package etcd provides ...
package etcd

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"time"

	"mysslee_qcloud/utils"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/coreos/etcd/pkg/transport"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var (
	dialTimeout       = 5 * time.Second
	requestTimeout    = 2 * time.Second
	testHookEndpoints = []string{"https://192.168.99.100:2379"}

	etcdConf Config
	EtcdCli  *MyEtcd
)

// EtcdConfig etcd config
type Config struct {
	Endpoints []string // etcd 服务器地址
	CACert    string
	Cert      string
	Key       string
	EchoURL   string // 服务器IP地址
}

type MyEtcd struct {
	etcd *clientv3.Client
}

// 初始化ETCD和注册App信息
func Init(appName AppName, conf Config) {
	etcdConf = conf
	var err error

	EtcdCli, err = GetEtcdCli()
	if err != nil {
		log.Infof("err:%v", err)
		return
	}
	if EtcdCli == nil {
		panic("没有获取到ectd的客户端")
		return
	}
	//连接 etcd
	go func() {
		defer utils.Recover(nil)
		ConnectETCD(appName)
	}()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	//关注app本身信息的变化
	go func() {
		defer utils.Recover(nil)
		EtcdCli.WatchDo(AppPrefix, AppCallback)
	}()

	LoadAppInfos(ctx, EtcdCli)
}

// Example:
// tlsInfo := transport.TLSInfo{
// 	CertFile:      "./tls-setup/certs/etcd2.pem",
// 	KeyFile:       "./tls-setup/certs/etcd2-key.pem",
// 	TrustedCAFile: "./tls-setup/certs/ca.pem",
// }
// tlsConfig, err := tlsInfo.ClientConfig()
// if err != nil {
// 	panic(err)
// }
// tlsConfig.InsecureSkipVerify = true
//
// http://127.0.0.1:2379, nil
// https://127.0.0.1:2379, tlsInfo
func Connect(endpoints []string) (*MyEtcd, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
	})
	if err != nil {
		return nil, err
	}

	return &MyEtcd{etcd: cli}, nil
}

// ca, cert, key path
func TLSConfig(trustCA, cert, key string) (*tls.Config, error) {
	return transport.TLSInfo{
		CertFile:      cert,
		KeyFile:       key,
		TrustedCAFile: trustCA,
	}.ClientConfig()
}

func ConnectAuth(endpoints []string, tlsConfig *tls.Config, username, password string) (*MyEtcd, error) {
	cli, err := clientv3.New(clientv3.Config{
		Endpoints:   endpoints,
		DialTimeout: dialTimeout,
		TLS:         tlsConfig,
		Username:    username,
		Password:    password,
	})
	if err != nil {
		return nil, err
	}

	return &MyEtcd{etcd: cli}, nil
}

// close
func (cli *MyEtcd) Close() error {
	return cli.etcd.Close()
}

// get current grpc connection
func (cli *MyEtcd) ActiveConnection() *grpc.ClientConn {
	return cli.etcd.ActiveConnection()
}

// watch the key when changed
// Use:
//	go cli.WatchDo(xxx, xxx)

func (cli *MyEtcd) WatchDoWithCallBack(key string, f func(clientv3.WatchResponse, func(int)), callback func(int)) {
	watchChan := cli.etcd.Watch(context.Background(), key, clientv3.WithPrefix())
	for wresp := range watchChan {
		f(wresp, callback)
	}
	// 测试etcd关掉，这里也不会结束，再启动etcd还能继续工作
	log.Warn("etcd watch end with key:", key)
}

func (cli *MyEtcd) WatchDo(key string, f func(clientv3.WatchResponse)) {
	watchChan := cli.etcd.Watch(context.Background(), key, clientv3.WithPrefix())
	for wresp := range watchChan {
		f(wresp)
	}
	// 测试etcd关掉，这里也不会结束，再启动etcd还能继续工作
	log.Warn("etcd watch end with key:", key)
}

// update ttl specified key
func (cli *MyEtcd) RenewLease(ctx context.Context, id clientv3.LeaseID) error {
	_, err := cli.etcd.KeepAliveOnce(ctx, id)
	if err != nil {
		return err
	}
	return nil
}

// put value
func (cli *MyEtcd) Put(ctx context.Context, key string, val string) error {
	_, err := cli.etcd.Put(ctx, key, val)
	if err != nil {
		return err
	}

	return nil
}

// put value with ttl
func (cli *MyEtcd) PutTTL(ctx context.Context, key, val string, ttl int64) (clientv3.LeaseID, error) {
	resp, err := cli.etcd.Grant(ctx, ttl)
	if err != nil {
		return 0, err
	}
	_, err = cli.etcd.Put(ctx, key, val, clientv3.WithLease(resp.ID))
	if err != nil {
		return 0, err
	}

	return resp.ID, nil
}

// get values
func (cli *MyEtcd) GetValue(ctx context.Context, key string) ([]*mvccpb.KeyValue, error) {
	resp, err := cli.etcd.Get(ctx, key, clientv3.WithPrefix())
	if err != nil {
		return nil, err
	}

	return resp.Kvs, nil
}

// get value with json
func (cli *MyEtcd) GetValueJSON(ctx context.Context, key string, result interface{}) error {
	resp, err := cli.etcd.Get(ctx, key)
	if err != nil {
		return err
	}

	return json.Unmarshal(resp.Kvs[0].Value, result)
}
