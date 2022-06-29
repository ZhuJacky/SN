// Package config provides ...
package config

import (
	"fmt"
	"html/template"
	"io/ioutil"
	"mysslee_qcloud/l5"
	"mysslee_qcloud/utils/etcd"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	yaml "gopkg.in/yaml.v2"
)

// run mode
const (
	MODE_DEV            = "DEV"
	MODE_TEST           = "TEST"
	MODE_PRE_PRODUCTION = "PRE_PRODUCTION"
	MODE_PRODUCTION     = "PRODUCTION"
)

// DatabaseConf the postgres config
type DatabaseConf struct {
	Driver string
	Source string
}

// RedisConf the redis config
type RedisConf struct {
	Address  string
	Password string
}

// PrometheusConf the prometheus config
type PrometheusConf struct {
	Gateway []string
	EchoURL string
	Open    bool
	Cert    string
	Key     string
}

// MysslConf MysslConf
type MysslConf struct {
	Domain         string
	Id             string
	Key            string
	AnaAPI         string
	DetectionCount int
}

// QCloud QCloud
type QCloud struct {
	AccountGateway string
	BillingGateway string
	NotifyGateway  string
	Module         string
	ThemeIds       []int
	Caller         string
	Env            string
}

// CKafkaSASL sasl配置
type CKafkaSASL struct {
	Username   string
	Password   string
	InstanceId string
}

// KafkaConf kafka配置
type KafkaConf struct {
	Topic           string
	SASL            CKafkaSASL
	Servers         []string
	ConsumerGroupId string
}

// PolarisBaseConf 北极星基础配置
type PolarisBaseConf struct {
	Service string
	Port    int
	Token   string
}

// PolarisConf 北极星配置
type PolarisConf struct {
	PolarisNamespace string
	Backend          PolarisBaseConf
	Checker          PolarisBaseConf
	Notifier         PolarisBaseConf
}

// AppConf the app config
type AppConf struct {
	RunMode    string
	ForceL5    bool
	OrmDebug   bool
	Database   DatabaseConf
	Redis      RedisConf
	Prometheus PrometheusConf
	Etcd       etcd.Config
	Kafka      KafkaConf
	Polaris    PolarisConf
	MySSL      MysslConf
	// DNSNodes   DNSNodes
	QCloud   QCloud
	Backend  BackendConf
	Checker  CheckerConf
	Notifier NotifierConf
	Region   string
}

// common params
var (
	Conf         = AppConf{}
	GMRootCAPath = "conf/gm_roots.pem"
	MailTmpl     *template.Template
	ProductPath  = "conf/product/product.json"
	NumCPU       = runtime.NumCPU()

	devPath     = "/data/release/newsslpod/"
	appconfPath = "conf/%s.yml"
	tmplPath    = "conf/tpl/*.tpl"
)

func init() {
	var prefix string
	// check dir
	_, err := os.Stat(devPath)
	if err == nil || !os.IsNotExist(err) {
		prefix = devPath
	} else {
		file, _ := exec.LookPath(os.Args[0])
		pwd, err := filepath.Abs(filepath.Dir(file))
		if err != nil {
			panic(err)
		}
		prefix = pwd + "/"
	}
	appconfPath = prefix + appconfPath
	GMRootCAPath = prefix + GMRootCAPath
	ProductPath = prefix + ProductPath

	// run mode
	var name string
	runMode := os.Getenv("RUN_MODE")
	if runMode == MODE_DEV {
		name = "dev"
	} else if runMode == MODE_TEST {
		name = "test"
	} else if runMode == MODE_PRE_PRODUCTION ||
		runMode == MODE_PRODUCTION {

		name = "prod"
	} else {
		panic("NotFound or incorrect env RUN_MODE")
	}
	Conf.RunMode = runMode

	// read file
	data, err := ioutil.ReadFile(fmt.Sprintf(appconfPath, name))
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &Conf)
	if err != nil {
		panic(err)
	}
	// L5
	forceL5 := os.Getenv("MYSSLEE_FORCE_L5") == "true"
	Conf.ForceL5 = forceL5
	if Conf.ForceL5 {
		l5.Init()
	}
	// DNS set
	// dns.SetServers(Conf.DNSNodes.Infos())

	// fix path
	Conf.Etcd.CACert = prefix + Conf.Etcd.CACert
	Conf.Etcd.Cert = prefix + Conf.Etcd.Cert
	Conf.Etcd.Key = prefix + Conf.Etcd.Key

	// parse email template
	tmplPath = prefix + tmplPath
	MailTmpl, err = template.ParseGlob(tmplPath)
	if err != nil {
		panic(err)
	}

	// prometheus
	Conf.Prometheus.Cert = prefix + Conf.Prometheus.Cert
	Conf.Prometheus.Key = prefix + Conf.Prometheus.Key

	// read env
	readEnv()
}

func readEnv() {
	// 读取CPU限制
	if os.Getenv("CPU_NUM") != "" {
		var err error
		NumCPU, err = strconv.Atoi(os.Getenv("CPU_NUM"))
		if err != nil {
			panic(err)
		}
	} else {
		panic("CPU_NUM is empty")
	}
	// region
	region := os.Getenv("SSLPOD_REGION")
	if region == "" {
		panic("SSLPOD_REGION env is empty")
	}
	Conf.Region = region
	// 直接用地域标识当作consumergroupid，不再从环境变量里读取
	Conf.Kafka.ConsumerGroupId = region
	// 配置修改为从七彩石直接拉取yml文件，不再从环境变量里读取，不要新增下面的逻辑了
	// 必须新增环境变量的，加在上面
	// kafka
	kafkaTopic := os.Getenv("KAFKA_TOPIC")
	kafkaUsername := os.Getenv("KAFKA_USERNAME")
	kafkaPassword := os.Getenv("KAFKA_PASSWORD")
	kafkaServer := os.Getenv("KAFKA_SERVER")
	// kafkaConsumerGroup := os.Getenv("KAFKA_CONSUMERGROUP")
	kafkaConsumerGroup := region
	if kafkaTopic != "" {
		Conf.Kafka.Topic = kafkaTopic
	}
	if kafkaUsername != "" {
		Conf.Kafka.SASL.Username = kafkaUsername
	}
	if kafkaPassword != "" {
		Conf.Kafka.SASL.Password = kafkaPassword
	}
	if kafkaServer != "" {
		Conf.Kafka.Servers = []string{kafkaServer}
	}
	if kafkaConsumerGroup != "" {
		Conf.Kafka.ConsumerGroupId = kafkaConsumerGroup
	}
	// polaris
	polarisNamespace := os.Getenv("POLARIS_NAMESPACE")
	if polarisNamespace != "" {
		Conf.Polaris.PolarisNamespace = polarisNamespace
		Conf.Polaris.Backend.Port, _ = strconv.Atoi(os.Getenv("POLARIS_BACKEND_PORT"))
		Conf.Polaris.Backend.Service = os.Getenv("POLARIS_BACKEND_SERVICE")
		Conf.Polaris.Backend.Token = os.Getenv("POLARIS_BACKEND_TOKEN")
		Conf.Polaris.Checker.Port, _ = strconv.Atoi(os.Getenv("POLARIS_CHECKER_PORT"))
		Conf.Polaris.Checker.Service = os.Getenv("POLARIS_CHECKER_SERVICE")
		Conf.Polaris.Checker.Token = os.Getenv("POLARIS_CHECKER_TOKEN")
		Conf.Polaris.Notifier.Port, _ = strconv.Atoi(os.Getenv("POLARIS_NOTIFIER_PORT"))
		Conf.Polaris.Notifier.Service = os.Getenv("POLARIS_NOTIFIER_SERVICE")
		Conf.Polaris.Notifier.Token = os.Getenv("POLARIS_NOTIFIER_TOKEN")
	}

	// postgres
	driver := os.Getenv("MYSSLEE_DB_DRIVER")
	source := os.Getenv("MYSSLEE_DB_SOURCE")
	if driver != "" && source != "" {
		Conf.Database.Driver = driver
		Conf.Database.Source = source
	}

	// redis
	addr := os.Getenv("MYSSLEE_REDIS_ADDRESS")
	password := os.Getenv("MYSSLEE_REDIS_PASSWORD")
	if addr != "" {
		Conf.Redis.Address = addr
		Conf.Redis.Password = password
	}

	// prometheus
	gateway := os.Getenv("MYSSLEE_PROMETHEUS_GATEWAY")
	echoURL := os.Getenv("MYSSLEE_PROMETHEUS_ECHOURL")
	if gateway != "" {
		Conf.Prometheus.Gateway = strings.Split(gateway, ",")
		Conf.Prometheus.EchoURL = echoURL
	}

	// etcd
	endpoints := os.Getenv("MYSSLEE_ETCD_ENDPOINTS")
	echoURL = os.Getenv("MYSSLEE_ETCD_ECHOURL")
	if endpoints != "" {
		Conf.Etcd.Endpoints = strings.Split(endpoints, ",")
		Conf.Etcd.EchoURL = echoURL
	}

	// myssl
	if domain := os.Getenv("MYSSLEE_MYSSL_DOMAIN"); domain != "" {
		Conf.MySSL.Domain = domain
	}
	if id := os.Getenv("MYSSLEE_MYSSL_ID"); id != "" {
		Conf.MySSL.Id = id
	}
	if key := os.Getenv("MYSSLEE_MYSSL_KEY"); key != "" {
		Conf.MySSL.Key = key
	}
	if api := os.Getenv("MYSSLEE_MYSSL_API"); api != "" {
		Conf.MySSL.AnaAPI = api
	}

	// qcloud
	if gateway := os.Getenv("MYSSLEE_QCLOUD_ACCOUNT_GATEWAY"); gateway != "" {
		Conf.QCloud.AccountGateway = gateway
	}
	if gateway := os.Getenv("MYSSLEE_QCLOUD_BILLING_GATEWAY"); gateway != "" {
		Conf.QCloud.BillingGateway = gateway
	}
	if gateway := os.Getenv("MYSSLEE_QCLOUD_NOTIFY_GATEWAY"); gateway != "" {
		Conf.QCloud.NotifyGateway = gateway
	}
	if module := os.Getenv("MYSSLEE_QCLOUD_MODULE"); module != "" {
		Conf.QCloud.Module = module
	}
	if themeIds := os.Getenv("MYSSLEE_QCLOUD_THEMEIDS"); themeIds != "" {
		for _, v := range strings.Split(themeIds, ",") {
			id, err := strconv.Atoi(v)
			if err != nil {
				panic(err)
			}
			Conf.QCloud.ThemeIds = append(Conf.QCloud.ThemeIds, id)
		}
	}
	if caller := os.Getenv("MYSSLEE_QCLOUD_CALLER"); caller != "" {
		Conf.QCloud.Caller = caller
	}
	if env := os.Getenv("MYSSLEE_QCLOUD_ENV"); env != "" {
		Conf.QCloud.Env = env
	}
}
