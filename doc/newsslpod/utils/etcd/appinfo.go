// Package etcd provides ...
package etcd

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"sort"
	"strings"
	"sync"

	"github.com/coreos/etcd/clientv3"
	"github.com/coreos/etcd/mvcc/mvccpb"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

var (
	AppInfos = make(map[AppName]map[string]MyAppInfo)
	//reportAppInfo = make(map[string]AppInfo)
	//adminAppInfo  = make(map[string]AppInfo)
	lock sync.RWMutex

	myNodeAddr     string
	myInstanceType string
)

type AppName string

const (
	AppPrefix = "/app/"

	// 获取机器信息的一些后缀
	SuffixInstallType = "/instance-type"
	SuffixLocalIPV4   = "/local-ipv4"
)

const (
	BackendApp  AppName = "/app/backend"
	CheckerApp  AppName = "/app/checker"
	NotifierApp AppName = "/app/notifier"
)

// MyAppInfo 向ETCD中注册的app的信息
type MyAppInfo struct {
	InstanceType string `json:"instance_type"`
	IntranetAddr string `json:"intranet_addr"`
	ExtranetAddr string `json:"extranet_addr"`
	JoinTime     string `json:"join_time"`
}

// GetInstanceType 获取当前实例机器类型
func GetInstanceType() (typ string, err error) {
	if myInstanceType == "" {
		resp, err := http.Get(etcdConf.EchoURL + SuffixInstallType)
		if err != nil {
			err = errors.New("etcd get my node instance type err: " + err.Error())
			return "", err
		}
		defer resp.Body.Close()

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			err = errors.New("etcd get my node instance type err: " + err.Error())
			return "", err
		}

		if len(buf) == 0 {
			err = errors.New("etcd get my node instance type empty")
			return "", err
		}

		myInstanceType = string(buf)
	}

	return myInstanceType, nil
}

// GetMyNodeAddr 获取当前实例内网地址
func GetMyNodeAddr() (addr string, err error) {
	if myNodeAddr == "" {
		resp, err := http.Get(etcdConf.EchoURL + SuffixLocalIPV4)
		if err != nil {
			err = errors.New("etcd get my node addr err: " + err.Error())
			return "", err
		}
		defer resp.Body.Close()
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil || len(data) == 0 {
			err = errors.New("etcd get my node addr err: " + err.Error())
		}
		myNodeAddr = string(data)
	}
	return myNodeAddr, nil
}

// /app/reportapp/172.31.1.123
func (a *AppName) FromKey(key []byte) bool {
	var pathIdx = 0
	for n, s := range key {
		if s == '/' {
			pathIdx++
		}
		if pathIdx == 3 {
			*a = AppName(key[:n])
			return true
		}
	}
	return false
}

func AppCallback(wresp clientv3.WatchResponse) {
	if err := wresp.Err(); err != nil {
		log.Error("etcd.CallbackInfo", err)
		return
	}

	var appName AppName

	for _, ev := range wresp.Events {
		switch ev.Type {
		case mvccpb.PUT:
			log.Info("etcd appinfo update: " + string(ev.Kv.Key))
			// 更新
			if !appName.FromKey(ev.Kv.Key) {
				return
			}
			addOrUpdateInfo(appName, ev.Kv.Key, ev.Kv.Value)
		case mvccpb.DELETE:
			log.Info("etcd appinfo delete: " + string(ev.Kv.Key))
			// 删除
			if !appName.FromKey(ev.Kv.Key) {
				return
			}
			deleteInfo(appName, ev.Kv.Key)
		}
	}
}

var GetAppInfo = GetAvailableIntranetAddr

//通过应用名称载入app 节点信息
func LoadAppInfos(ctx context.Context, cli *MyEtcd) error {
	lock.Lock()
	defer lock.Unlock()

	kvs, err := cli.GetValue(ctx, string(AppPrefix))
	if err != nil {
		return err
	}

	for _, kv := range kvs {
		var appName AppName
		if !appName.FromKey(kv.Key) {
			return errors.New("appname key invalid")
		}

		appNodeInfos, exist := AppInfos[appName]
		if !exist {
			appNodeInfos = make(map[string]MyAppInfo)
		}

		info := MyAppInfo{}
		err = json.Unmarshal(kv.Value, &info)
		if err != nil {
			log.Error("etcd.LoadAppInfos json.Unmarshal ", err)
			continue
		}

		appNodeInfos[string(kv.Key)] = info
		AppInfos[appName] = appNodeInfos
	}

	log.Info("etcd appinfo load current appinfos: ", len(AppInfos))
	return nil
}

func GetAppInfoByAppName(appName AppName) []MyAppInfo {
	lock.RLock()
	defer lock.RUnlock()

	var apps []MyAppInfo
	for _, v := range AppInfos[appName] {
		apps = append(apps, v)
	}
	sort.Slice(apps, func(i, j int) bool {
		return strings.Compare(apps[i].IntranetAddr, apps[j].IntranetAddr) > 0
	})
	return apps
}

func addOrUpdateInfo(appName AppName, key, value []byte) {
	lock.Lock()
	defer lock.Unlock()

	appInfo, exist := AppInfos[appName]
	if !exist {
		appInfo = make(map[string]MyAppInfo)
	}

	nodeId := string(key)
	info, ok := appInfo[nodeId]
	if !ok {
		err := json.Unmarshal(value, &info)
		if err != nil {
			log.Error("etcd.appInfoUpdate json.Unmarshal ", err)
			return
		}
	}
	appInfo[nodeId] = info
	AppInfos[appName] = appInfo
}

func deleteInfo(appName AppName, key []byte) {
	lock.Lock()
	defer lock.Unlock()

	appInfo, exist := AppInfos[appName]
	if !exist {
		log.Errorf("删除etcd信息时，应用名为:%v的信息不存在", appName)
		return
	}

	delete(appInfo, string(key))
}
