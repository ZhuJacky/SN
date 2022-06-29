// Package etcd provides ...
package etcd

import (
	"fmt"
	"testing"
	"time"

	"github.com/coreos/etcd/clientv3"
	log "github.com/sirupsen/logrus"
)

var mycli *MyEtcd

func init() {
	testHookEndpoints = []string{"http://192.168.252.135:2379"}

	tlsConfig, err := TLSConfig("./tls-setup/certs/ca.pem", "./tls-setup/certs/etcd1.pem", "./tls-setup/certs/etcd1-key.pem")
	if err != nil {
		log.Panicf("err:%v", err)
	}
	mycli, err = ConnectAuth(testHookEndpoints, tlsConfig, "", "")
	if err != nil {
		log.Errorf("a err:%v", err.Error())
	}
}

func TestWatchDo(t *testing.T) {
	f := func(wresp clientv3.WatchResponse) {
		if err := wresp.Err(); err != nil {
			t.Error(err)
		}
		for _, ev := range wresp.Events {
			fmt.Println(ev.Type, ev.Kv.String(), ev.PrevKv.String())
		}
	}

	time.AfterFunc(time.Second*20, func() { mycli.etcd.Close() })
	mycli.WatchDo("test", f)
	time.Sleep(time.Second)
}
