// Package timer provides ...
package timer

import (
	"bytes"
	"encoding/gob"
	"encoding/json"
	"errors"
	"fmt"
	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/prom"
	"mysslee_qcloud/config"
	"mysslee_qcloud/kafka"
	"mysslee_qcloud/model"
	"mysslee_qcloud/polaris"
	"mysslee_qcloud/redis"
	"net/http"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Init TODO
func Init() {
	gob.Register([]model.DomainResult{})
	go needCheckDomains()
}

// 设计思考：
// 1. 预估有 25w 域名（包括多IP）
// 2. 有 1 台backend节点
// 3. 有 5 台checker节点
//
// 每节点每5s需要分发任务 25w/10/60*3=1250
func needCheckDomains() {
	// 检测个数公式
	limit := func(total int) int {
		if total == 0 {
			return 0
		}

		length := 1
		ips, err := redis.ScanApp(redis.BackendApp)
		if err != nil {
			logrus.Error("redis.ScanApp: ", err)
		} else if len(ips) > 0 {
			length = len(ips)
		}
		count := int(float64(total/length/10/60) * config.Conf.Backend.Task.Duration)
		if count >= 1250 {
			logrus.Error("check domains in 5s overload 1250")
			return 1250
		}
		return count + 20
	}
	t := time.NewTicker(time.Second * time.Duration(config.Conf.Backend.Task.Interval))
	for now := range t.C {
		// 同一个时刻只有一个timer运行, 5秒后自动解锁
		if !redis.Lock("sslpod:backend:timer") {
			continue
		}
		// 获取当前需要检查的域名个数
		// hei := time.Now()
		drs, err := db.GetNeedCheckDomains(now, limit)

		// 推送至kafka进行全地域检测
		if os.Getenv("PushKafka") == "on" {
			go needCheckDomainsToKafka(drs)
		}

		if err != nil {
			logrus.Error("needCheckDomains.GetNeedCheckDomains: ", err)
			continue
		}
		logrus.Debugf("get %d needCheckDomains\n", len(drs))
		if l := len(drs); l == 0 {
			continue
		}
		// send task, rand checker app
		ips, err := redis.ScanApp(redis.CheckerApp)
		if err != nil || len(ips) == 0 {
			logrus.Error("needCheckDomains.CheckerApp: not found checker app")
			continue
		}
		l := len(ips)
		total := len(drs)
		size := total/l + 1

		buf := &bytes.Buffer{}
		// 平均分配到每个app
		for i, ip := range ips {
			buf.Reset()

			start := i * size
			if start > total-1 {
				break
			}
			end := (i + 1) * size
			if end > total {
				end = total
			}
			err = gob.NewEncoder(buf).Encode(drs[start:end])
			if err != nil {
				logrus.Error("needCheckDomains.Encode: ", err)
				continue
			}
			// NOTE
			now2 := time.Now()
			req, err := http.NewRequest(http.MethodPost, "http://"+ip+":20010/api/task", buf)
			if err != nil {
				logrus.Error("needCheckDomains.NewRequest: ", err)
				continue
			}
			logrus.Info("node ", ip, time.Now().Sub(now2).Seconds())
			// basicauth
			req.SetBasicAuth("backend", "2fde7b35d23015f3e94462f3880dfb4d")
			// prometheus
			prom.PromTaskDispatch.WithLabelValues(ip, "total").
				Add(float64(size))
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				// prometheus
				prom.PromTaskDispatch.WithLabelValues(ip, "failed").
					Add(float64(size))

				logrus.Error("needCheckDomains.Post: ", err)
				continue
			}
			resp.Body.Close()
		}
	}
}

// DistributorStatus 任务分配器状态
var (
	DistributorStatus string

	indexChecker = 0
	defaultDur   = time.Second * 5
	errOverload  = errors.New("check domains in 5s overload 3580")
)

const (
	stautsNormal   = "正常"
	statusOverload = "超载"
)

// 获取需要检测的域名，推至kafka
func needCheckDomainsToKafka(drs []model.DomainResult) {
	// 检测在北极星有健康的checker时，才下发
	checkers, err := polaris.GetInstanceByService("checker")
	if err != nil {
		return
	}
	var hasHealthyChecker = false
	for _, i := range checkers {
		if i.IsHealthy() == true {
			hasHealthyChecker = true
			break
		}
	}
	if hasHealthyChecker == false {
		return
	}

	var kafkaDomainInfos = []model.KafkaDomainInfo{}
	var domainIDs = []int{}
	for _, i := range drs {
		domainIDs = append(domainIDs, i.Id)
	}
	domainIPInfos := db.GetDomainIps(domainIDs)
	for _, i := range drs {
		temp := model.KafkaDomainInfo{
			DomainID:       i.Id,
			Domain:         i.Domain,
			PunyCodeDomain: i.PunyCodeDomain,
			ServerType:     i.ServerType,
			IP:             i.IP,
			Port:           i.Port,
			DomainFlag:     i.DomainFlag,
		}
		if j, ok := domainIPInfos[i.Id]; ok {
			temp.IPPorts = j
			// 自动检测永远打开，显示的时候再根据客户开没开启自动检测进行过滤
			temp.IsAutoDetect = true
		} else {
			temp.IsAutoDetect = true
		}
		kafkaDomainInfos = append(kafkaDomainInfos, temp)
	}
	if l := len(kafkaDomainInfos); l == 0 {
		return
	}
	// send task to kafka
	total := len(kafkaDomainInfos)
	batch := 10
	offset := 0
	for offset < total {

		start := offset
		if start > total-1 {
			break
		}
		end := offset + batch
		if end > total {
			end = total
		}
		kafkaDomainInfoString, _ := json.Marshal(kafkaDomainInfos[start:end])
		kafka.ProduceMessage(kafkaDomainInfoString)
		offset = offset + batch
	}
	fmt.Println("push to kafka")
}
