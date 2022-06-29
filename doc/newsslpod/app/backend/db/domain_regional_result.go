package db

import (
	"encoding/json"
	"mysslee_qcloud/common"
	"mysslee_qcloud/model"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
)

// GetDomainIps 获得域名设置的ip表
// key domainID value []*model.IPPort  域名-域名对应的所有uin手动监控的ip
func GetDomainIps(domainIDs []int) map[int][]model.IPPort {
	ret := []*model.DomainIps{}
	rret := map[int][]model.IPPort{}
	err := gormDB.Model(&model.DomainIps{}).Where("domain_id in (?)", domainIDs).Find(&ret).Error
	if err != nil {
		return rret
	}
	ipDuplicateCheck := map[int]map[string]struct{}{}
	for _, i := range ret {
		ipports := []model.IPPort{}
		json.Unmarshal([]byte(i.IpPorts), &ipports)
		// 去重
		for _, j := range ipports {
			if _, ok := ipDuplicateCheck[i.DomainID][j.IP+":"+j.Port]; ok {
				continue
			} else {
				rret[i.DomainID] = append(rret[i.DomainID], j)
			}
		}
	}
	return rret
}

// GetDomainIPsByUin 根据账号获得域名IP
func GetDomainIPsByUin(domainID int, uin string) (*model.DomainIps, error) {
	ret := &model.DomainIps{}
	err := gormDB.Model(&model.DomainIps{}).Where("domain_id=?", domainID).Where("uin=?", uin).First(ret).Error
	return ret, err
}

// GetDomainAllRegionalResult 根据域名ID和地域获得
func GetDomainAllRegionalResult(domainID int) ([]*model.DomainRegionalResult, error) {
	ret := []*model.DomainRegionalResult{}
	err := gormDB.Model(&model.DomainRegionalResult{}).Where("domain_id = ?", domainID).
		Find(&ret).Error
	return ret, err
}

// SaveDomainIPs 保存域名IP
func SaveDomainIPs(d *model.DomainIps) error {
	err := gormDB.Model(&model.DomainIps{}).Save(d).Error
	return err
}

// GetDomainAllRegionalResultsWithUIN 获取域名的全地域的检测信息，并根据uin的关注ip限制进行ip过滤
func GetDomainAllRegionalResultsWithUIN(uin string, domainId int, domain string) ([]map[string]interface{}, bool) {
	// 先直接获得域名全地域的检测结果
	r, _ := GetDomainAllRegionalResult(domainId)
	// 再获得该账号该域名的IP信息
	var staticIPs = &model.DomainIps{}
	staticIPs, err = GetDomainIPsByUin(domainId, uin)
	// 如果该账号该域名的IP信息为空，则只返回自动检测的
	if err == gorm.ErrRecordNotFound {
		staticIPs = &model.DomainIps{
			IsAutoDetect: true,
		}
	}
	// 整理用户的固定IP和端口 整理为a[ip+port] = []的形式
	var staticIPPorts = []model.IPPort{}
	err = json.Unmarshal([]byte(staticIPs.IpPorts), &staticIPPorts)
	var staticIPSet = map[string]struct{}{}
	if err != nil {
		logrus.Error("GetDomainAllRegionalResultsWithUIN.UnmarshalIpPorts: ", err)
	}
	for _, ip := range staticIPPorts {
		staticIPSet[ip.IP+":"+ip.Port] = struct{}{}
	}
	res := []map[string]interface{}{}
	// 判断检测结果里的ip是否应该返回给客户
	for _, rr := range r {
		var dr []model.DetectionResult
		err := json.Unmarshal([]byte(rr.DetectionResult), &dr)
		if err != nil {
			logrus.Error("GetDomainAllRegionalResultsWithUIN.UnmarshalDetectionResult: ", err)
		}
		j := 0
		for _, drr := range dr {
			// 如果客户开启了自动检测，那么自动检测到的ip直接通过
			if staticIPs.IsAutoDetect && drr.IsAuto {
				dr[j] = drr
				j++
				continue
			}
			// 其他情况下把客户自选ip都加上，不在自选ip内的不加，即可
			if _, ok := staticIPSet[drr.IP+":"+drr.Port]; ok {
				dr[j] = drr
				j++
				continue
			}
			// 兼容监控的是ip的情况，如果域名和ip相等那么直接添加
			if drr.IP == domain {
				dr[j] = drr
				j++
				continue
			}
		}
		res = append(res, map[string]interface{}{
			"Region":            rr.Region,
			"DetectionResult":   dr[:j],
			"LastDetectionTime": rr.LastDetectionTime,
		})
	}
	return res, staticIPs.IsAutoDetect
}

// GetDomainAllIPCerts 返回一个域名的全地域聚合去重后的证书信息+ip+是否开启自动检测
func GetDomainAllIPCerts(uin string, domainId int, domain string) ([]map[string]interface{}, []map[string]interface{},
	bool, error) {
	a, isAutoDetect := GetDomainAllRegionalResultsWithUIN(uin, domainId, domain)
	// 记录每一个证书hash对应的ip，hash做key ip做value
	hashIPs := map[string]string{}
	// 记录每一个ip的状态
	ipStatus := map[string][]string{}
	// 记录每一个证书的状态 hash做key status是value
	hashCertStatus := map[string]string{}
	lastdt := ""
	for _, i := range a {
		dr := i["DetectionResult"].([]model.DetectionResult)
		lastdt = i["LastDetectionTime"].(time.Time).Format("2006-01-02 15:04:05")
		for _, drr := range dr {
			for _, hash := range drr.Hashes {
				// 去重
				if !strings.Contains(hashIPs[hash], drr.IP+":"+drr.Port) {
					hashIPs[hash] = hashIPs[hash] + "," + drr.IP + ":" + drr.Port
					hashCertStatus[hash] = drr.Status
				}
			}
			ipStatus[drr.IP+":"+drr.Port] = append(ipStatus[drr.IP+":"+drr.Port], drr.Status)
		}
	}
	// 处理证书列表
	r := []map[string]interface{}{}
	for hash, ips := range hashIPs {
		certInfo, _ := GetCertInfoByHash(hash)
		r = append(r, map[string]interface{}{
			"CertStatus":       common.NewCertStatusText[hashCertStatus[hash]],
			"CommonName":       certInfo.CN,
			"Brand":            certInfo.Brand,
			"EncryptAlgorithm": certInfo.KeyAlgo,
			"CertHash":         certInfo.Hash,
			"DNSNames":         certInfo.SANs,
			"Issuer":           certInfo.Issuer,
			"CertBeginTime":    certInfo.BeginTime,
			"CertEndTime":      certInfo.EndTime,
			"IPs":              ips,
		})
	}
	// 处理ip列表, 兼容部分异常的情况
	ipr := []map[string]interface{}{}
	for ipport, status := range ipStatus {
		s := ""
		haveNormal := false
		haveAbnormal := false
		for _, s := range status {
			if s == common.CertTrust {
				haveNormal = true
			} else {
				haveAbnormal = true
			}
		}
		if haveNormal && haveAbnormal {
			s = common.CertPartAbnormal
		}
		if haveNormal && !haveAbnormal {
			s = common.CertTrust
		}
		if !haveNormal && haveAbnormal {
			s = status[0]
		}
		ipr = append(ipr, map[string]interface{}{
			"IP":                strings.Split(ipport, ":")[0],
			"Port":              strings.Split(ipport, ":")[1],
			"Status":            s,
			"LastDetectionTime": lastdt,
		})
	}
	return r, ipr, isAutoDetect, nil
}
