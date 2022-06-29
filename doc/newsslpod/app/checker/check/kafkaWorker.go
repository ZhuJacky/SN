package check

import (
	"bytes"
	"context"
	"encoding/gob"
	"encoding/json"
	"mysslee_qcloud/app/checker/db"
	"mysslee_qcloud/common"
	"mysslee_qcloud/config"
	"mysslee_qcloud/core/myconn"
	"mysslee_qcloud/dns"
	"mysslee_qcloud/model"
	"mysslee_qcloud/polaris"
	"mysslee_qcloud/utils"
	"net/http"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/idna"
)

// kafka任务检测
type kafkaCheckWorker struct {
	busy     int32
	workChan chan *model.KafkaDomainInfo
}

func (w *kafkaCheckWorker) incr() {
	atomic.AddInt32(&w.busy, 1)
}

func (w *kafkaCheckWorker) decr() {
	atomic.AddInt32(&w.busy, -1)
}

func (w *kafkaCheckWorker) do() {
	for dr := range w.workChan {
		// select {}
		w.kafkaCheck(dr)
		w.decr()
	}
}

// kafka任务的检测
func (w *kafkaCheckWorker) kafkaCheck(result *model.KafkaDomainInfo) {
	var (
		err      error
		callback = model.CallbackToBackend{
			IPCerts: map[string]model.CertWithErr{},
			Err:     "",
		}
		ips     []string // save ips
		ipPorts = map[string]model.IPPort{}
	)

	// 自动获取IPv4
	if result.IsAutoDetect && !utils.ValidateIP(result.Domain) {
		ips, err = dns.LookupHost(context.Background(), result.PunyCodeDomain)
		for _, ip := range ips {
			if _, ok := ipPorts[ip]; !ok {
				ipPorts[ip] = model.IPPort{
					IP:   ip,
					Port: result.Port,
				}
			}
		}
	}

	// 将自动获取和手动的拼装
	if len(result.IPPorts) != 0 {
		for _, i := range result.IPPorts {
			if _, ok := ipPorts[i.IP]; !ok {
				ipPorts[i.IP] = i
			}
		}
	}
	// 将原有的也加进去
	if result.DomainFlag != 0 {
		if _, ok := ipPorts[result.IP]; !ok {
			ipPorts[result.IP] = model.IPPort{
				IP:   result.IP,
				Port: result.Port,
			}
		}
	}
	var wg sync.WaitGroup
	wg.Add(len(ipPorts))
	certsChan := make(chan *model.CertWithErr, 10)

	now := time.Now()
	var ipCount = 0
	for _, ipport := range ipPorts {
		ipCount++
		if ipCount > 10 {
			return
		}
		go func(it model.IPPort) {
			defer wg.Done()
			cert := model.CertWithErr{}
			cert.Cert, err = singleIPCheck(result.PunyCodeDomain, it.Port, it.IP, result.ServerType)
			if err != nil {
				cert.Err = err.Error()
			}
			cert.IPorts = it
			certsChan <- &cert
		}(ipport)
	}
	wg.Wait()
	close(certsChan)
	domain, _ := idna.ToUnicode(result.Domain)
	logrus.Println("cost:", domain, time.Now().Sub(now), len(ipPorts))

	for c := range certsChan {
		callback.IPCerts[c.IPorts.IP] = *c
	}
	buf := &bytes.Buffer{}
	err = gob.NewEncoder(buf).Encode(callback)
	if err != nil {
		logrus.Error("kafka check gob.NewEncoder fail", err, callback)
		return
	}
	resultHash := utils.SHA1(buf.Bytes())
	dbRegionResult, err := db.GetDomainRegionalResult(result.DomainID, config.Conf.Region)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			dbRegionResult = nil
		} else {
			logrus.Error("kafka check db.GetDomainRegionalResult fail ", err)
			return
		}
	}
	if dbRegionResult != nil && dbRegionResult.ResultHash == resultHash {
		db.UpdateDomainRegionalResultDetectionTime(result.DomainID, config.Conf.Region)
		return
	}
	// 将自动获取的IP搞成hash，用来索引
	autoIPs := make(map[string]struct{})
	for _, value := range ips {
		autoIPs[value] = struct{}{}
	}

	oldDcs, err := db.GetDomainCert(result.DomainID)
	var infos = []*model.CertInfo{}
	var detectionHashes = []*model.DetectionResult{}
	if callback.Err == "" {
		for ip, certInfos := range callback.IPCerts {
			// infos = append(infos, certInfos.Cert...)
			// 得去掉重复的证书
			infos = addCertInfosAviodDuplicate(infos, certInfos.Cert)
			if certInfos.Err == "" && len(certInfos.Cert) != 0 {
				detectionHashes = append(detectionHashes, &model.DetectionResult{
					IP:     ip,
					Port:   certInfos.IPorts.Port,
					Status: certInfos.Cert[0].TrustStatus,
					Hashes: func(certinfos []*model.CertInfo) []string {
						var certhashes []string
						for _, cert := range certinfos {
							certhashes = append(certhashes, cert.Hash)
						}
						return certhashes
					}(certInfos.Cert),
					IsAuto: isContainIP(ip, autoIPs),
				})
			} else {
				detectionHashes = append(detectionHashes, &model.DetectionResult{
					IP:     ip,
					Port:   certInfos.IPorts.Port,
					Status: common.CannotConnect,
					Hashes: []string{},
					IsAuto: isContainIP(ip, autoIPs),
				})
			}
		}
	}
	// infos 如果没有数据，则跳过更新
	if len(infos) != 0 {
		// 证书是否改变
		delDcs, upDcs, newDcs, needAggr := diffDomainCert(result.DomainID, infos, oldDcs)
		for _, info := range infos {
			if !db.IsExistCertInfo(info.Hash) {
				err = db.AddCertInfo(info)
				if err != nil {
					logrus.Error("kafkaCheck.AddCertInfo ", err)
					return
				}
			}
		}

		// 证书发生变化
		if len(delDcs)+len(newDcs) > 0 || needAggr {
			err = db.UpDomainCert(result.DomainID, delDcs, upDcs, newDcs)
			if err != nil {
				logrus.Error("kafkaCheck.UpDomainCert ", err)
				return
			}
		}
	}
	detectionResultJSON, err := json.Marshal(detectionHashes)

	regionalDBParam := &model.DomainRegionalResult{
		DomainID:          result.DomainID,
		Region:            config.Conf.Region,
		DetectionResult:   string(detectionResultJSON),
		ResultHash:        resultHash,
		LastDetectionTime: time.Now().Local(),
	}
	if dbRegionResult != nil {
		err = db.UpdateDomainRegionalResult(regionalDBParam)
	} else {
		err = db.InsertDomainRegionalResult(regionalDBParam)
	}
}

func sendCallbackToBackend(callback *model.CallbackToBackend) {
	buf := &bytes.Buffer{}
	err := gob.NewEncoder(buf).Encode(callback)
	if err != nil {
		logrus.Error("kafka check gob.NewEncoder fail", err, callback)
		return
	}
	backendIns, err := polaris.GetInstanceByService("backend")
	var backendIP = ""
	if len(backendIns) != 0 {
		backendIP = backendIns[0].GetHost()
	}
	req, err := http.NewRequest(http.MethodPost, "http://"+backendIP+":20000/api/taskCallback", buf)
	if err != nil {
		logrus.Error("check.kafkaCheck.NewCallbackRequest: ", err)
		return
	}
	var backendAc, backendPw string
	for a, p := range config.Conf.Backend.BasicAuth {
		backendAc, backendPw = a, p
	}
	req.SetBasicAuth(backendAc, backendPw)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		logrus.Error("kafkaCheck.PostCallback: ", err)
	}
	resp.Body.Close()
}

func singleIPCheck(punyCodeDomain string, port string, ip string, serverType int) ([]*model.CertInfo, error) {
	mailDirect := false
	// TLS 通用调整为 HTTPS 检测
	switch serverType {
	case 1, 2, 3:
		mailDirect = true
	case 4:
		serverType = 0
	}
	params := &myconn.CheckParams{
		Domain:     punyCodeDomain,
		Port:       port,
		Ip:         ip,
		ServerType: myconn.ServerType(serverType),
		MailDirect: mailDirect,
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Conf.Checker.Task.FastTimeout)*time.Second)
	defer cancel()

	infos, err := GenerateMultipleCertificates(ctx, params)
	// 暂时只返回第一张sni证书，去除多余的证书
	if len(infos) != 0 {
		infos = infos[0:1]
	}
	return infos, err
}

func isContainIP(ip string, autoIPs map[string]struct{}) bool {
	if _, ok := autoIPs[ip]; ok {
		return true
	}
	return false
}

func addCertInfosAviodDuplicate(infos []*model.CertInfo, newInfos []*model.CertInfo) []*model.CertInfo {
	var certhashes = map[string]struct{}{}
	for _, i := range infos {
		if _, ok := certhashes[i.Hash]; ok {
			continue
		}
		certhashes[i.Hash] = struct{}{}
	}
	for _, i := range newInfos {
		if _, ok := certhashes[i.Hash]; ok {
			continue
		}
		infos = append(infos, i)
		certhashes[i.Hash] = struct{}{}
	}
	return infos
}
