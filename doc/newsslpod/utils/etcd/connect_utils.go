package etcd

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"time"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/etcdserver/api/v3rpc/rpctypes"
	"github.com/pkg/errors"
	"github.com/prometheus/client_golang/prometheus"
	log "github.com/sirupsen/logrus"
)

var (
	PromEtcdError = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "etcd_connected_error",
			Help: "the count of etcd error",
		},
	)
)

// ConnectETCD 连接etcd的帮助类
func ConnectETCD(appName AppName) {
	var err error
	defer func() {
		if err != nil {
			time.AfterFunc(time.Second, func() {
				ConnectETCD(appName)
			})
		}
	}()

	addr, err := GetMyNodeAddr()
	if err != nil {
		log.Errorf("load myappinfo err: %v", err)
		return
	}

	instanceType, err := GetInstanceType()
	if err != nil {
		log.Errorf("load my instance type err: %v", err)
		return
	}

	cli, err := GetEtcdCli()
	if err != nil {
		log.Errorf("get etcdcli err: %v", err)
		return
	}

	var appInfo = &MyAppInfo{
		IntranetAddr: addr,
		InstanceType: instanceType,
	}

	id, err := regAppInfo(cli, appName, appInfo)
	if err != nil {
		log.Error("regAppInfo err:", err)
		return
	}
	time.AfterFunc(time.Second*5, func() { renewAppInfoLease(cli, id, appName, appInfo) })
}

func GetEtcdCli() (cli *MyEtcd, err error) {
	tlsConfig, err := TLSConfig(etcdConf.CACert, etcdConf.Cert, etcdConf.Key)
	if err != nil {
		return nil, err
	}
	tlsConfig.InsecureSkipVerify = true

	return ConnectAuth(etcdConf.Endpoints, tlsConfig, "", "")
}

//通过域名获取etcd服务
func GetEtcdCliByDomain(domain string, etcdConfig *Config) (cli *MyEtcd, err error) {
	tlsConfig, err := TLSConfig(etcdConfig.CACert, etcdConfig.Cert, etcdConfig.Key)
	if err != nil {
		return nil, err
	}
	tlsConfig.InsecureSkipVerify = true
	ips, err := net.LookupHost(domain)
	if err != nil || len(ips) == 0 {
		return nil, errors.New("没有查询到etcd的ip")
	}
	return ConnectAuth([]string{"https://" + ips[0] + ":2379"}, tlsConfig, "", "")

}

func regAppInfo(cli *MyEtcd, appName AppName, appInfo *MyAppInfo) (id clientv3.LeaseID, err error) {
	var key = string(appName) + "/" + appInfo.IntranetAddr
	log.Infof("register etcd appinfo with " + key)

	appInfo.JoinTime = time.Now().String()
	data, err := json.Marshal(appInfo)
	if err != nil {
		return 0, errors.New(fmt.Sprintf("appinfo marshal err: %v", err))
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	id, err = cli.PutTTL(ctx, key, string(data), 12)
	return
}

//更新租期
func renewAppInfoLease(cli *MyEtcd, id clientv3.LeaseID, appName AppName, appInfo *MyAppInfo) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	err := cli.RenewLease(ctx, id)
	if err != nil {
		PromEtcdError.Inc()
		if err == rpctypes.ErrLeaseNotFound {
			appInfo.JoinTime = time.Now().String()
			id, err = regAppInfo(cli, appName, appInfo)
			if err != nil {
				log.Errorf("regAppInfo2 err:%v", err)
			}
		} else {
			log.Errorf("etcd.RenewLease err:%v", err)
		}
	}
	time.AfterFunc(time.Second*3, func() { renewAppInfoLease(cli, id, appName, appInfo) })
}
