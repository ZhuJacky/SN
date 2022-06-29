package polaris

import (
	"errors"
	"fmt"
	"mysslee_qcloud/config"
	"net"
	"strings"
	"sync/atomic"
	"time"

	"git.code.oa.com/polaris/polaris-go/api"
	"git.code.oa.com/polaris/polaris-go/pkg/model"
	"github.com/sirupsen/logrus"
)

var (
	consumer  api.ConsumerAPI
	provider  api.ProviderAPI
	namespace string
	seconds   int
)

func init() {
	var err error
	consumer, err = api.NewConsumerAPI()
	if err != nil {
		panic(err)
	}
	provider = api.NewProviderAPIByContext(consumer.SDKContext())
	provider.SDKContext().GetConfig()
	namespace = config.Conf.Polaris.PolarisNamespace
	api.SetLoggersLevel(api.NoneLog)
}

// GetInstanceByServiceAndRegion 根据服务名和区域获得实例
func GetInstanceByServiceAndRegion(service string, region string) {
	var flowID uint64
	getInstancesReq := &api.GetOneInstanceRequest{}
	getInstancesReq.FlowID = atomic.AddUint64(&flowID, 1)
	getInstancesReq.Namespace = namespace
	getInstancesReq.Service = service
	getInstancesReq.SourceService = &model.ServiceInfo{
		Namespace: namespace,
		Service:   service,
		Metadata: map[string]string{
			"region": region,
		},
	}
	getInstResp, err := consumer.GetOneInstance(getInstancesReq)
	if err != nil {
		logrus.Error("GetInstanceByServiceAndRegion fail", err)
	}
	fmt.Println(getInstResp)
}

// RegisterBackend 注册backend
func RegisterBackend() {
	request := &api.InstanceRegisterRequest{}
	request.Namespace = namespace
	request.Service = config.Conf.Polaris.Backend.Service
	request.ServiceToken = config.Conf.Polaris.Backend.Token
	request.Host = GetIP()
	request.Port = config.Conf.Polaris.Backend.Port
	request.Metadata = map[string]string{
		"region": config.Conf.Region,
	}
	request.SetTTL(2)
	resp, err := provider.Register(request)
	if nil != err {
		panic(fmt.Sprintf("fail to register instance, err %v", err))
	}
	logrus.Info("success to register instance, id is %s", resp.InstanceID)
	go HeartBeat(request)
}

// RegisterChecker 注册checker
func RegisterChecker() {
	request := &api.InstanceRegisterRequest{}
	request.Namespace = namespace
	request.Service = config.Conf.Polaris.Checker.Service
	request.ServiceToken = config.Conf.Polaris.Checker.Token
	request.Host = GetIP()
	request.Port = config.Conf.Polaris.Checker.Port
	request.Metadata = map[string]string{
		"region": config.Conf.Region,
	}
	request.SetTTL(2)
	resp, err := provider.Register(request)
	if nil != err {
		panic(fmt.Sprintf("fail to register instance, err %v", err))
	}
	logrus.Info("success to register instance, id is %s", resp.InstanceID)
	go HeartBeat(request)
}

// RegisterNotifier 注册notifier
func RegisterNotifier() {
	request := &api.InstanceRegisterRequest{}
	request.Namespace = namespace
	request.Service = config.Conf.Polaris.Notifier.Service
	request.ServiceToken = config.Conf.Polaris.Notifier.Token
	request.Host = GetIP()
	request.Port = config.Conf.Polaris.Notifier.Port
	request.SetTTL(2)
	resp, err := provider.Register(request)
	if nil != err {
		panic(fmt.Sprintf("fail to register instance, err %v", err))
	}
	logrus.Info("success to register instance, id is %s", resp.InstanceID)
	go HeartBeat(request)
}

// HeartBeat 上报心跳
func HeartBeat(request *api.InstanceRegisterRequest) {
	hbRequest := &api.InstanceHeartbeatRequest{}
	hbRequest.Namespace = request.Namespace
	hbRequest.Service = request.Service
	hbRequest.Host = request.Host
	hbRequest.Port = request.Port
	hbRequest.ServiceToken = request.ServiceToken
	for {
		if err := provider.Heartbeat(hbRequest); nil != err {
			logrus.Error("fail to heartbeat, error is ", err)
		}
		<-time.After(1500 * time.Millisecond)
	}
}

// GetInstanceByService 根据服务名获得实例
func GetInstanceByService(service string) ([]model.Instance, error) {
	if service == "backend" {
		service = config.Conf.Polaris.Backend.Service
	} else if service == "checker" {
		service = config.Conf.Polaris.Checker.Service
	} else if service == "notifier" {
		service = config.Conf.Polaris.Notifier.Service
	} else {
		return nil, errors.New("invalid service")
	}
	var flowID uint64
	getInstancesReq := &api.GetOneInstanceRequest{}
	getInstancesReq.FlowID = atomic.AddUint64(&flowID, 1)
	getInstancesReq.Namespace = namespace
	getInstancesReq.Service = service
	getInstancesReq.SourceService = &model.ServiceInfo{
		Namespace: namespace,
		Service:   service,
	}
	getInstResp, err := consumer.GetOneInstance(getInstancesReq)
	if err != nil {
		logrus.Error("GetInstanceByService fail", err)
	}
	if getInstResp == nil {
		return []model.Instance{}, err
	}
	return getInstResp.Instances, err
}

// GetAllInstanceByService 根据服务名获得实例
func GetAllInstanceByService(service string) ([]model.Instance, error) {
	if service == "backend" {
		service = config.Conf.Polaris.Backend.Service
	} else if service == "checker" {
		service = config.Conf.Polaris.Checker.Service
	} else if service == "notifier" {
		service = config.Conf.Polaris.Notifier.Service
	} else {
		return nil, errors.New("invalid service")
	}
	var flowID uint64
	getInstancesReq := &api.GetAllInstancesRequest{}
	getInstancesReq.FlowID = atomic.AddUint64(&flowID, 1)
	getInstancesReq.Namespace = namespace
	getInstancesReq.Service = service
	getInstResp, err := consumer.GetAllInstances(getInstancesReq)
	if err != nil {
		logrus.Error("GetAllInstanceByService fail", err)
	}
	return getInstResp.Instances, err
}

// GetIP 获得本机ip
func GetIP() string {
	conn, err := net.Dial("udp", "119.29.29.29:53")
	if err != nil {
		panic(err)
	}
	localAddr := conn.LocalAddr().(*net.UDPAddr)
	ip := strings.Split(localAddr.String(), ":")[0]
	return ip
}
