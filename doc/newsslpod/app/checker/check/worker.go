// Package check provides ...
package check

import (
	"context"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"mysslee_qcloud/app/checker/db"
	"mysslee_qcloud/app/checker/prom"
	"mysslee_qcloud/brand"
	"mysslee_qcloud/common"
	"mysslee_qcloud/config"
	"mysslee_qcloud/core"
	"mysslee_qcloud/core/cert"
	"mysslee_qcloud/core/myconn"
	"mysslee_qcloud/core/ocsp"
	"mysslee_qcloud/dns"
	"mysslee_qcloud/limiter"
	"mysslee_qcloud/model"
	"mysslee_qcloud/utils"
	"mysslee_qcloud/utils/certutils"
	"net/http"
	"strings"
	"sync/atomic"
	"time"

	"github.com/sirupsen/logrus"
)

// UserAgent UserAgent
const UserAgent = "MySSL-EE/1.0"

// 快速检测
type fastCheckWorker struct {
	busy     int32
	workChan chan *model.DomainResult
}

func (w *fastCheckWorker) incr() {
	atomic.AddInt32(&w.busy, 1)
}

func (w *fastCheckWorker) decr() {
	atomic.AddInt32(&w.busy, -1)
}

func (w *fastCheckWorker) do() {
	for dr := range w.workChan {
		// select {}
		needFullCheck := w.fastCheck(dr)
		w.decr()

		if needFullCheck {
			DomainChecker.DoFull(dr)
		}
	}
}

// 快速检测
func (w *fastCheckWorker) fastCheck(result *model.DomainResult) (needFullCheck bool) {
	var (
		trustStatus string
		err         error
		infos       []*model.CertInfo
		fields      = make(map[string]interface{}) // update fields
	)
	needFullCheck = false
	if time.Now().UTC().Sub(result.LastFullDetectionTime.UTC()) > 24*time.Hour {
		if result.LastFullDetectionTime.Equal(model.TimeZeroAt) ||
			limiter.LimitMap.Load("fullCheck", config.Conf.MySSL.DetectionCount, time.Second*40).Get() {
			result.LastFullDetectionTime = time.Now().UTC()
			needFullCheck = true

			fields["last_full_detection_time"] = result.LastFullDetectionTime
		} else {
			logrus.Info("ratelimit fullcheck: ", result.Domain)
		}
	}

	// 重新获取IPv4
	if result.IP == "" || result.DomainFlag&model.DomainFlagBindIP == 0 {
		ips, err := dns.LookupHost(context.Background(), result.PunyCodeDomain)
		if err != nil || len(ips) == 0 {
			_ = fmt.Errorf("not found ip or %v", err)
			result.IP = ""
		} else if result.IP != ips[0] {
			result.IP = ips[0]

			fields["ip"] = result.IP
		}
	}
	serverType := result.ServerType
	mailDirect := false
	// TLS 通用调整为 HTTPS 检测
	switch serverType {
	case 1, 2, 3:
		mailDirect = true
	case 4:
		serverType = 0
	}
	params := &myconn.CheckParams{
		Domain:     result.PunyCodeDomain,
		Port:       result.Port,
		Ip:         result.IP,
		ServerType: myconn.ServerType(serverType),
		MailDirect: mailDirect,
	}
	ctx, cancel := context.WithTimeout(context.Background(),
		time.Duration(config.Conf.Checker.Task.FastTimeout)*time.Second)
	defer cancel()
	now := time.Now()
	infos, err = GenerateMultipleCertificates(ctx, params)
	logrus.Println("cost:", result.Domain, time.Now().Sub(now))

	if err != nil {
		prom.PromFastDetection.WithLabelValues("failed").Inc()

		trustStatus = common.CannotConnect

		// 连接异常不进行完整的检测
		needFullCheck = false
		if result.ResultHash != "BF21A9E8FBC5A3846FB05B4FA0859E0917B2202F" {
			logrus.Error("[fastCheck.GenerateMultipleCertificates] ", err)

			result.FullDetectionResult = []byte("{}")
			result.Brand = "unknown"
			result.Grade = common.SecureLevelUnknown
			status := model.CalculateStatusForUnknown()
			data, _ := json.Marshal(status)
			result.DomainStatus = string(data)
			result.ResultHash = "BF21A9E8FBC5A3846FB05B4FA0859E0917B2202F"

			fields["full_detection_result"] = result.FullDetectionResult
			fields["brand"] = result.Brand
			fields["grade"] = result.Grade
			fields["domain_status"] = result.DomainStatus
			fields["result_hash"] = result.ResultHash
		}
	} else {
		trustStatus = infos[0].TrustStatus

		oldDcs, err := db.GetDomainCert(result.Id)
		if err != nil {
			logrus.Error("fastCheck.GetDomainCert ", err)
			return
		}
		// 证书是否改变 暂时只取第一张sni证书，忽略其他证书
		delDcs, upDcs, newDcs, needAggr := diffDomainCert(result.Id, infos[0:1], oldDcs)
		for _, info := range infos[0:1] {
			if !db.IsExistCertInfo(info.Hash) {
				err = db.AddCertInfo(info)
				if err != nil {
					logrus.Error("fastCheck.AddCertInfo ", err)
					return
				}
			}
		}
		// 检测正常
		if result.Brand != infos[0].Brand {
			result.Brand = infos[0].Brand
			needAggr = true

			fields["brand"] = result.Brand
		}
		// 证书发生变化，聚合
		if len(delDcs)+len(newDcs) > 0 || needAggr {
			data, _ := json.Marshal(model.CalculateStatusCert(result.DomainStatus, infos))
			result.DomainStatus = string(data)

			fields["domain_status"] = string(data)

			err = db.UpDomainCert(result.Id, delDcs, upDcs, newDcs)
			if err != nil {
				logrus.Error("fastCheck.UpDomainCert ", err)
				return
			}
			err = db.SetAccountAggrFlag(result.Id)
			if err != nil {
				logrus.Error("fullCheck.SetAccountAggrFlag: ", err)
			}
		}
	}
	// 更新 prev_status
	if result.PrevStatus != result.TrustStatus {
		result.PrevStatus = result.TrustStatus

		fields["prev_status"] = result.PrevStatus
	}
	// 更新 trust_status
	if result.TrustStatus != trustStatus {
		result.TrustStatus = trustStatus

		fields["trust_status"] = result.TrustStatus

		// 设置聚合flag
		err = db.SetAccountAggrFlag(result.Id)
		if err != nil {
			logrus.Error("fullCheck.SetAccountAggrFlag: ", err)
		}
	}
	// 如果之前检测是无法连接，后一次检测是非无法连接，需要重新完整检测
	if result.PrevStatus != result.TrustStatus &&
		result.PrevStatus == common.CannotConnect {
		needFullCheck = true
	}

	// 如何证书不可信
	if result.TrustStatus != common.CertTrust {
		// 两次状态不一致，更新 account_domain 关系表
		if needNotify(result) {
			db.ResetAccountDomainNoticedAt(result.Id)
		}
		go WarnNotice(result)
	}
	// 更新数据
	if len(fields) > 0 {
		db.UpdateDomainResult(result.Id, fields)
	}
	return
}

// 全量检测
type fullCheckWorker struct {
	busy     int32
	workChan chan *model.DomainResult
}

func (w *fullCheckWorker) incr() {
	atomic.AddInt32(&w.busy, 1)
}

func (w *fullCheckWorker) decr() {
	atomic.AddInt32(&w.busy, -1)
}

func (w *fullCheckWorker) do() {
	for dr := range w.workChan {
		w.fullCheck(dr, 5*time.Second, 0)
		w.decr()
	}
}

func (w *fullCheckWorker) fullCheck(result *model.DomainResult, after time.Duration, count int) {
	// time.Sleep(time.Second * 30)
	// logrus.Info("mock test fullcheck")
	// return
	// 异步检测
	finish, data, report, err := syncCheck(result)
TIMEOUT:
	if err != nil {
		prom.PromFullDetection.WithLabelValues("failed").Inc()

		result.Grade = common.SecureLevelUnknown
		// 如果无法检测，重新设置成空
		result.FullDetectionResult = []byte("{}")
		// status := model.CalculateStatusForUnknown()
		// statusData, _ := json.Marshal(status)
		// result.DomainStatus = string(statusData)

		// "{}" 的hash, 是否需要聚合
		if result.ResultHash == "BF21A9E8FBC5A3846FB05B4FA0859E0917B2202F" {
			return
		}
		result.ResultHash = "BF21A9E8FBC5A3846FB05B4FA0859E0917B2202F"
		db.UpdateDomainResult(result.Id, map[string]interface{}{
			"grade":                 result.Grade,
			"full_detection_result": result.FullDetectionResult,
			// "domain_status":         result.DomainStatus,
			"result_hash": result.ResultHash,
		})
		// 设置聚合flag
		err = db.SetAccountAggrFlag(result.Id)
		if err != nil {
			logrus.Error("fullCheck.SetAccountAggrFlag: ", err)
		}
	} else if finish {
		result.FullDetectionResult = data
		result.Grade = report.Data.Basic.LevelStr
		status := model.CalculateStatusAll(result.DomainStatus, report.Data)
		statusData, _ := json.Marshal(status)
		result.DomainStatus = string(statusData)

		// 计算结果hash
		h := utils.SHA1(data)
		if h == result.ResultHash {
			return
		}
		result.ResultHash = h
		db.UpdateDomainResult(result.Id, map[string]interface{}{
			"grade":                 result.Grade,
			"full_detection_result": result.FullDetectionResult,
			"domain_status":         result.DomainStatus,
			"result_hash":           result.ResultHash,
		})
		// 设置聚合flag
		err = db.SetAccountAggrFlag(result.Id)
		if err != nil {
			logrus.Error("fullCheck.SetAccountAggrFlag: ", err)
		}
	} else {
		if count >= config.Conf.Checker.Task.FullTimeout/5 {
			err = errors.New("Full check timeout")
			goto TIMEOUT
		} else {
			// 继续获取结果
			count++
			time.AfterFunc(after, func() {
				w.fullCheck(result, after, count)
			})
		}
	}
}

func syncCheck(result *model.DomainResult) (finish bool, fullResult []byte, report *model.MessageForUnmarshal,
	err error) {
	var params []utils.KV
	params = []utils.KV{
		{Key: "partnerId", Value: config.Conf.MySSL.Id},
		{Key: "timestamp", Value: time.Now().Unix()},
		{Key: "expire", Value: 200},
		{Key: "f", Value: "1"},
		{Key: "count", Value: 0},
		{Key: "domain", Value: result.PunyCodeDomain},
		{Key: "port", Value: result.Port},
		{Key: "ip", Value: result.IP},
	}
	plaintext, signed := utils.SignReport(config.Conf.MySSL.Key, params)
	urlStr := config.Conf.MySSL.AnaAPI + "?" + plaintext + "&signature=" + signed
	client := &http.Client{
		Timeout: 2 * time.Minute, // 二分钟超时
	}

	req, err := http.NewRequest(http.MethodGet, urlStr, nil)
	if err != nil {
		logrus.Errorf("创建请求错误：%v", err)
		return
	}
	req.Header.Set("User-Agent", UserAgent)

	resp, err := client.Do(req)
	if err != nil {
		logrus.Errorf("发送request错误：%v", err)
		return
	}
	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		logrus.Errorf("读取数据response body错误：%v", err)
		return
	}

	if resp.StatusCode != http.StatusOK {
		logrus.Infof("domain :%v,port :%v full_check_error: %v", result.Domain, result.Port, string(data))
		return
	}

	report = &model.MessageForUnmarshal{}
	err = json.Unmarshal(data, &report)
	if err != nil {
		logrus.Errorf("反解析错误:%v", err)
		return
	}

	if report.Progress == model.ProgressDone {
		if report.Data == nil || report.Data.Basic == nil {
			return true, nil, nil, errors.New("myssl数据返回无效")
		}
		return true, data, report, nil
	} else if report.Progress == model.ProgressErr {
		return true, nil, nil, errors.New("myssl无法完成检测")
	}

	return false, nil, nil, nil
}

// 判断是否需要进行通知
func needNotify(result *model.DomainResult) bool {
	// 第一次检测
	if result.PrevStatus == "" {
		return true
	} else if result.TrustStatus != result.PrevStatus { // 是否更糟糕
		return isWorseThanLast(result.TrustStatus, result.PrevStatus)
	}
	return false
}

// 判断是否比上一次情况还要差
func isWorseThanLast(trustStatus, prevStatus string) bool {
	trustCode := common.ChangeStatusToCode(trustStatus)
	prevCode := common.ChangeStatusToCode(prevStatus)
	if trustCode == common.UnknownStatusCode ||
		prevCode == common.UnknownStatusCode {
		return false
	}
	return trustCode < prevCode
}

// GenerateMultipleCertificates 获取证书，使用 myssl.com 的数据证书品牌
func GenerateMultipleCertificates(ctx context.Context, params *myconn.CheckParams) ([]*model.CertInfo, error) {
	if params.Ip == "" {
		return nil, fmt.Errorf("not found ip: %s", params.Domain)
	}

	certNodes, err := core.GetMultipleCertInfo(ctx, params)
	if err != nil {
		return nil, err
	}

	var infos []*model.CertInfo
	// 优先使用从 SNI 拿取到的证书
	for _, v := range certNodes {
		if v.Cert.SNI && len(v.Cert.CertsInfo) > 0 {
			info := convertToModelCert(ctx, v)
			infos = append(infos, info)
		}
	}
	// 非 SNI
	for _, v := range certNodes {
		if v.Cert.SNI {
			continue
		}
		if len(v.Cert.CertsInfo) > 0 {
			info := convertToModelCert(ctx, v)
			infos = append(infos, info)
		}
	}
	if len(infos) == 0 {
		return nil, errors.New("not found cert info")
	}
	return infos, nil
}

func convertToModelCert(ctx context.Context, node *core.CertNode) *model.CertInfo {
	certInfo := node.Cert.CertsInfo[0]

	typ := cert.GetAuditType(certInfo.X509)
	info := &model.CertInfo{
		Hash:        certInfo.Sha1,
		SN:          certInfo.SN,
		CN:          node.Cert.CN,
		SANs:        strings.Join(node.Cert.DNSS, ","),
		O:           certInfo.Organization,
		OU:          certInfo.OrganizationUnit,
		Street:      strings.Join(node.Cert.ServerCertificates[0].Subject.StreetAddress, ","),
		City:        strings.Join(node.Cert.ServerCertificates[0].Subject.Locality, ","),
		Province:    strings.Join(node.Cert.ServerCertificates[0].Subject.Province, ","),
		Country:     strings.Join(node.Cert.ServerCertificates[0].Subject.Country, ","),
		KeyAlgo:     fmt.Sprintf("%v %v bits", certInfo.KeyType, certInfo.KeyLong),
		SignAlgo:    certInfo.SignAlgo,
		CertType:    cert.CertTypeToString(typ),
		BeginTime:   certInfo.NotBefore,
		Issuer:      certInfo.Issuer,
		EndTime:     certInfo.NotAfter,
		TrustStatus: GetTrustStatus(IsTrust(node.Cert)),
		RawPEM:      certInfo.CertPem,
	}
	brand, _ := brand.GetCertBrand(ctx, []*x509.Certificate{certInfo.X509})
	info.Brand = brand
	return info
}

// IsTrust IsTrust
func IsTrust(c *core.OutCerts) int {
	var status int
	info := c.CertsInfo[0]
	if c.TrustStatus == cert.Untrusted {
		status |= CertUntrust
	}

	if c.TrustStatus == cert.BlackList {
		status |= CertInBlack
	}

	if !c.DomainInCert {
		status |= CertNameUnmatch
	}
	if certutils.IsExpired(info.X509) {
		status |= CertExpired
	}
	if c.OCSP.Status == ocsp.Revoked || c.OCSPStaplingInfo.Status == ocsp.Revoked {
		status |= CertRevoke
	}
	if info.SignAlgo == x509.MD2WithRSA || info.SignAlgo == x509.MD5WithRSA || info.SignAlgo == x509.SHA1WithRSA ||
		info.SignAlgo == x509.ECDSAWithSHA1 {
		status |= CertUseWeakSignAlgo
	}
	days := int(info.NotAfter.UTC().Sub(time.Now()).Hours()) / 24
	if days >= 0 && days <= 7 {
		status |= CertExpiring7
	}
	if days > 7 && days < 30 {
		status |= CertExpiring30
	}
	return status
}

// 需要处理无法连接的问题
const (
	CertTrust           = iota
	CertExpiring30      = 1 << (iota - 1)
	CertExpiring7       = 1 << (iota - 1)
	CertUseWeakSignAlgo = 1 << (iota - 1)
	CertUntrust         = 1 << (iota - 1)
	CertNameUnmatch     = 1 << (iota - 1)
	CertInBlack         = 1 << (iota - 1)
	CertRevoke          = 1 << (iota - 1)
	CertExpired         = 1 << (iota - 1)
)

// GetTrustStatus 获取可信状态的中文表述
func GetTrustStatus(status int) string {
	if status&CertExpired != 0 {
		return common.CertExpired
	} else if status&CertRevoke != 0 {
		return common.CertRevoke
	} else if status&CertInBlack != 0 {
		return common.CertBlackList
	} else if status&CertNameUnmatch != 0 {
		return common.CertNotMatch
	} else if status&CertUntrust != 0 {
		return common.CertUntrusted
	} else if status&CertUseWeakSignAlgo != 0 {
		return common.CertUseWeakKey
	} else if status&CertExpiring7 != 0 {
		return common.CertExpiring7
	} else if status&CertExpiring30 != 0 {
		return common.CertExpiring30
	} else {
		return common.CertTrust
	}
}

// 网站证书是否改变
func isCertChanged(infos []*model.CertInfo, oldDcs []*model.DomainCert) bool {
	if len(infos) != len(oldDcs) {
		return true
	}
	for _, v := range infos {
		for _, v2 := range oldDcs {
			if v.Hash != v2.Hash {
				return true
			}
		}
	}
	return false
}

// 证书域名关系
func diffDomainCert(domainId int, infos []*model.CertInfo,
	oldDcs []*model.DomainCert) (delDcs, upDcs, newDcs []*model.DomainCert, needAggr bool) {

	for _, info := range infos {
		new := true
		for _, dc := range oldDcs {
			if dc.Hash == info.Hash {
				new = false
				if dc.TrustStatus != info.TrustStatus {
					dc.TrustStatus = info.TrustStatus
					needAggr = true
				}
				upDcs = append(upDcs, dc)
				break
			}
		}
		if new {
			newDcs = append(newDcs, &model.DomainCert{
				DomainId:    domainId,
				TrustStatus: info.TrustStatus,
				Hash:        info.Hash,
			})
		}
	}
	for _, dc := range oldDcs {
		del := true
		for _, v := range upDcs {
			if dc.Hash == v.Hash {
				del = false
				break
			}
		}
		if del {
			delDcs = append(delDcs, dc)
		}
	}
	return
}
