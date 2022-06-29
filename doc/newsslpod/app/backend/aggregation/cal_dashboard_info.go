package aggregation

import (
	"encoding/json"
	"math"
	"time"

	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/common"
	"mysslee_qcloud/model"
)

func CalDashboardInfo(accountId int) error {
	//获取用户所有关注的域名
	var (
		messages  []*model.MessageForUnmarshal
		certInfos = make(map[int][]*model.CertInfo)
	)

	ids, err := db.GetDomainsIdByAccountId(accountId)
	if err != nil {
		return err
	}
	for _, id := range ids {
		result, err := db.GetDomainResultById(id)
		if err != nil {
			return err
		}
		msg := &model.MessageForUnmarshal{}
		err = json.Unmarshal(result.FullDetectionResult, msg)
		if err != nil {
			return err
		}
		messages = append(messages, msg)
		// 查找证书
		infos, err := db.GetDomainCertInfo(id)
		if err != nil {
			return err
		}
		certInfos[id] = infos
	}
	// 全量检测，聚合
	securityLevel, err := calSecurityLevel(messages)
	if err != nil {
		return err
	}
	sslBugs, err := calBugs(messages)
	if err != nil {
		return err
	}
	compliance, err := calCompliance(messages)
	if err != nil {
		return err
	}

	// 快速检测，聚合
	certBrands, err := calCertBrand2(certInfos)
	if err != nil {
		return err
	}
	certValidTime, err := calValidity2(certInfos)
	if err != nil {
		return err
	}
	certTypes, err := calCertType2(certInfos)
	if err != nil {
		return err
	}

	dashboard := &model.DashboardShow{
		SecurityLevelPie:         securityLevel,
		CertBrandsPie:            certBrands,
		CertValidTimePie:         certValidTime,
		CertTypePie:              certTypes,
		SSLBugsLoopholeHistogram: sslBugs,
		ComplianceHistogram:      compliance,
	}

	data, err := json.Marshal(dashboard)
	if err != nil {
		return err
	}

	err = db.UpdateDashBoardResultByUid(accountId, string(data))
	return err
}

//计算安全评级分布
func calSecurityLevel(messages []*model.MessageForUnmarshal) (securityLevel []*common.NameValue, err error) {
	var (
		levelAPlus   = 0
		levelA       = 0
		levelACut    = 0
		levelB       = 0
		levelC       = 0
		levelD       = 0
		levelE       = 0
		levelF       = 0
		levelT       = 0
		levelUnknown = 0
	)

	if len(messages) == 0 {
		return nil, nil
	}

	//计算数量
	for _, msg := range messages {
		if msg.Data == nil {
			levelUnknown++
			continue
		}

		if msg.Progress == "done" { //只处理能正常检测的
			switch msg.Data.Basic.Level {
			case common.LevelAPlus:
				levelAPlus++
			case common.LevelA:
				levelA++
			case common.LevelACut:
				levelACut++
			case common.LevelB:
				levelB++
			case common.LevelC:
				levelC++
			case common.LevelD:
				levelD++
			case common.LevelE:
				levelE++
			case common.LevelF:
				levelF++
			case common.LevelT:
				levelT++
			}
		}
	}

	//填充数据

	if levelAPlus != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelAPlus,
			levelAPlus,
		})
	}

	if levelA != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelA,
			levelA,
		})
	}

	if levelACut != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelA,
			levelA,
		})
	}

	if levelB != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelB,
			levelB,
		})
	}

	if levelC != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelC,
			levelC,
		})
	}

	if levelD != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelD,
			levelD,
		})
	}

	if levelE != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelA,
			levelA,
		})
	}
	if levelF != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelF,
			levelF,
		})
	}

	if levelT != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			common.SecureLevelT,
			levelT,
		})
	}

	if levelUnknown != 0 {
		securityLevel = append(securityLevel, &common.NameValue{
			"未知",
			levelUnknown,
		})
	}
	return securityLevel, nil
}

// 计算证书品牌分布2
func calCertBrand2(certInfos map[int][]*model.CertInfo) (brands []*common.NameValue, err error) {
	if len(certInfos) == 0 {
		return nil, nil
	}
	var unknownCount int
	for _, infos := range certInfos {
		if len(infos) == 0 {
			unknownCount++
			continue
		}
		for _, info := range infos {
			name := findNameValue(brands, info.Brand)
			if name == nil { //map中没有该品牌的信息
				brands = append(brands, &common.NameValue{
					info.Brand,
					1,
				})
			} else { //map中有该品牌的信息
				name.Value++
			}
		}
	}
	if unknownCount != 0 {
		brands = append(brands, &common.NameValue{
			"未知",
			unknownCount,
		})
	}
	return brands, nil
}

//计算证书品牌分布
func calCertBrand(messages []*model.MessageForUnmarshal) (brands []*common.NameValue, err error) {
	if len(messages) == 0 {
		return nil, nil
	}

	var unknownCount int
	for _, msg := range messages {
		if msg.Data == nil {
			unknownCount++
			continue
		}
		if msg.Progress == "done" {
			sniCerts, notSniCerts := model.DistinguishCerts(msg.Data)
			//如果有通过sni获取的证书，那么将屏蔽非sni获取的证书
			if len(sniCerts) > 0 { //有通过sni获取的证书
				for _, cert := range sniCerts {
					name := findNameValue(brands, cert.BrandName)
					if name == nil { //map中没有该品牌的信息
						brands = append(brands, &common.NameValue{
							cert.BrandName,
							1,
						})
					} else { //map中有该品牌的信息
						name.Value++
					}
				}
			} else { //没有通过sni获取的证书
				for _, cert := range notSniCerts {
					name := findNameValue(brands, cert.BrandName)
					if name == nil {
						brands = append(brands, &common.NameValue{
							cert.BrandName,
							1,
						})
					} else {
						name.Value++
					}
				}
			}
		}
	}

	if unknownCount != 0 {
		brands = append(brands, &common.NameValue{
			"未知",
			unknownCount,
		})
	}
	return brands, nil
}

//计算证书有效期分布2
func calValidity2(certInfos map[int][]*model.CertInfo) (validity []*common.NameValue, err error) {
	if len(certInfos) == 0 {
		return nil, nil
	}

	var unknownCount int
	for _, infos := range certInfos {
		if len(infos) == 0 {
			unknownCount++
			continue
		}
		for _, info := range infos {
			days := math.Ceil(info.EndTime.UTC().Sub(time.Now().UTC()).Hours() / 24)

			scope := common.ValidityGte90
			if days <= 0 {
				scope = common.ValidityLt0
			} else if days < 30 {
				scope = common.ValidityLt30
			} else if days < 60 {
				scope = common.ValidityLt60
			} else if days < 90 {
				scope = common.ValidityLt90
			}
			name := findNameValue(validity, scope)
			if name == nil {
				validity = append(validity, &common.NameValue{
					scope,
					1,
				})
			} else {
				name.Value++
			}
		}
	}
	if unknownCount != 0 {
		validity = append(validity, &common.NameValue{
			"未知",
			unknownCount,
		})
	}
	return validity, nil
}

//计算证书有效期分布
func calValidity(messages []*model.MessageForUnmarshal) (validity []*common.NameValue, err error) {
	if len(messages) == 0 {
		return nil, nil
	}

	var unknownCount int

	for _, msg := range messages {
		if msg.Data == nil {
			unknownCount++
			continue
		}
		if msg.Progress == "done" {
			sniCerts, notSniCerts := model.DistinguishCerts(msg.Data)
			//同样区分sni
			if len(sniCerts) > 0 {
				for _, cert := range sniCerts {
					scope := model.GetCertValidaty(cert)
					name := findNameValue(validity, scope)
					if name == nil {
						validity = append(validity, &common.NameValue{
							scope,
							1,
						})
					} else {
						name.Value++
					}
				}
			} else {
				for _, cert := range notSniCerts {
					scope := model.GetCertValidaty(cert)
					name := findNameValue(validity, scope)
					if name == nil {
						validity = append(validity, &common.NameValue{
							scope,
							1,
						})
					} else {
						name.Value++
					}
				}
			}

		}
	}

	if unknownCount != 0 {
		validity = append(validity, &common.NameValue{
			"未知",
			unknownCount,
		})
	}
	return validity, nil
}

//计算证书类型分布2
func calCertType2(certInfos map[int][]*model.CertInfo) (certTypes []*common.NameValue, err error) {
	if len(certInfos) == 0 {
		return nil, nil
	}
	var unknownCount int
	for _, infos := range certInfos {
		if len(infos) == 0 {
			unknownCount++
			continue
		}
		for _, info := range infos {
			if info.CertType == "NoAudit" {
				unknownCount++
				continue
			}
			name := findNameValue(certTypes, info.CertType)
			if name == nil {
				certTypes = append(certTypes, &common.NameValue{
					info.CertType,
					1,
				})
			} else {
				name.Value++
			}
		}
	}
	if unknownCount != 0 {
		certTypes = append(certTypes, &common.NameValue{
			"未知",
			unknownCount,
		})
	}
	return certTypes, nil
}

//计算证书类型分布
func calCertType(messages []*model.MessageForUnmarshal) (certTypes []*common.NameValue, err error) {
	if len(messages) == 0 {
		return nil, nil
	}
	var unknownCount int
	for _, msg := range messages {
		if msg.Data == nil {
			unknownCount++
			continue
		}
		if msg.Progress == "done" {
			sniCerts, notSniCerts := model.DistinguishCerts(msg.Data)
			if len(sniCerts) > 0 {
				for _, cert := range sniCerts {
					certType := cert.CertType
					if certType == "NoAudit" {
						unknownCount++
						continue
					}
					name := findNameValue(certTypes, certType)
					if name == nil {
						certTypes = append(certTypes, &common.NameValue{
							certType,
							1,
						})
					} else {
						name.Value++
					}
				}
			} else {
				for _, cert := range notSniCerts {
					certType := cert.CertType
					if certType == "NoAudit" {
						unknownCount++
						continue
					}
					name := findNameValue(certTypes, certType)
					if name == nil {
						certTypes = append(certTypes, &common.NameValue{
							certType,
							1,
						})
					} else {
						name.Value++
					}
				}
			}
		}
	}

	if unknownCount != 0 {
		certTypes = append(certTypes, &common.NameValue{
			"未知",
			unknownCount,
		})
	}
	return certTypes, nil
}

//计算ssl漏洞分布
func calBugs(messages []*model.MessageForUnmarshal) (bugs []*common.NameValueChildren, err error) {
	var (
		drownSupport         int
		cssSupport           int
		heartbleedSupport    int
		paddingOracleSupport int
		tlsPoodleSupport     int
		freakSupport         int
		logjamSupport        int
		poodleSupport        int
		crimeSupport         int
		//robotDetectSupport   int
		unknownCount int
		total        int
	)

	// 8 种漏洞;
	total = len(messages)
	if total == 0 {
		return nil, nil
	}
	bugs = make([]*common.NameValueChildren, 8)

	for _, msg := range messages {
		if msg.Data == nil {
			unknownCount++
			continue
		}
		if msg.Progress == "done" {
			//drown漏洞的数量
			if msg.Data.Bugs.Drown.Support != 0 {
				drownSupport++
			}

			//css漏洞的数量
			if msg.Data.Bugs.CCS.Support != 0 {
				cssSupport++
			}
			// heartbleed漏洞的数据
			if msg.Data.Bugs.Heartbleed.Support != 0 {
				heartbleedSupport++
			}

			//paddingOracle漏洞的数据
			if msg.Data.Bugs.PaddingOracle.Support != 0 {
				paddingOracleSupport++
			}

			//tlspoodle漏洞的数量
			if msg.Data.Bugs.TLSPOODLE.Support != 0 {
				tlsPoodleSupport++
			}

			//freak漏洞的数量
			if msg.Data.Bugs.FREAK.Support != 0 {
				freakSupport++
			}

			//logjam漏洞的数量
			if msg.Data.Bugs.Logjam.Support != 0 {
				logjamSupport++
			}

			if msg.Data.Bugs.POODLE.Support != 0 {
				poodleSupport++
			}

			if msg.Data.Bugs.CRIME.Support != 0 {
				crimeSupport++
			}
			//if msg.Data.Bugs.RobotDetect.Support != 0 {
			//	robotDetectSupport++
			//}
		}
	}

	//drown漏洞
	drown := make([]*common.NameValue, 3)
	drown[0] = &common.NameValue{
		common.BugsAffect,
		drownSupport,
	}
	drown[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	drown[2] = &common.NameValue{
		common.BugsUnaffect,
		total - drownSupport - unknownCount,
	}
	bugs[0] = &common.NameValueChildren{
		common.BugDrown,
		drown,
	}

	//paddingOracle 漏洞
	paddingOracle := make([]*common.NameValue, 3)
	paddingOracle[0] = &common.NameValue{
		common.BugsAffect,
		paddingOracleSupport,
	}
	paddingOracle[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	paddingOracle[2] = &common.NameValue{
		common.BugsUnaffect,
		total - paddingOracleSupport - unknownCount,
	}
	bugs[1] = &common.NameValueChildren{
		common.BugOpenSSLPaddingOracle,
		paddingOracle,
	}

	//freak 漏洞
	freak := make([]*common.NameValue, 3)
	freak[0] = &common.NameValue{
		common.BugsAffect,
		freakSupport,
	}
	freak[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	freak[2] = &common.NameValue{
		common.BugsUnaffect,
		total - freakSupport - unknownCount,
	}
	bugs[2] = &common.NameValueChildren{
		common.BugFreak,
		freak,
	}

	//logjam 漏洞
	logjam := make([]*common.NameValue, 3)
	logjam[0] = &common.NameValue{
		common.BugsAffect,
		logjamSupport,
	}
	logjam[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	logjam[2] = &common.NameValue{
		common.BugsUnaffect,
		total - logjamSupport - unknownCount,
	}
	bugs[3] = &common.NameValueChildren{
		common.BugFreak,
		logjam,
	}

	//css 漏洞
	css := make([]*common.NameValue, 3)
	css[0] = &common.NameValue{
		common.BugsAffect,
		cssSupport,
	}
	css[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	css[2] = &common.NameValue{
		common.BugsUnaffect,
		total - cssSupport - unknownCount,
	}
	bugs[4] = &common.NameValueChildren{
		common.BugOpenSSLCCS,
		css,
	}

	// heartbleed 漏洞
	heartbleed := make([]*common.NameValue, 3)
	heartbleed[0] = &common.NameValue{
		common.BugsAffect,
		heartbleedSupport,
	}
	heartbleed[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	heartbleed[2] = &common.NameValue{
		common.BugsUnaffect,
		total - heartbleedSupport - unknownCount,
	}
	bugs[5] = &common.NameValueChildren{
		common.BugHeartbleed,
		heartbleed,
	}

	//poodle漏洞
	poodle := make([]*common.NameValue, 3)
	poodle[0] = &common.NameValue{
		common.BugsAffect,
		poodleSupport,
	}
	poodle[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	poodle[2] = &common.NameValue{
		common.BugsUnaffect,
		total - poodleSupport - unknownCount,
	}
	bugs[6] = &common.NameValueChildren{
		common.BugPOODLE,
		poodle,
	}

	//crime漏洞
	crime := make([]*common.NameValue, 3)
	crime[0] = &common.NameValue{
		common.BugsAffect,
		crimeSupport,
	}
	crime[1] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	crime[2] = &common.NameValue{
		common.BugsUnaffect,
		total - crimeSupport - unknownCount,
	}
	bugs[7] = &common.NameValueChildren{
		common.BugCRIME,
		crime,
	}

	return bugs, nil
}

//计算合规性分布
func calCompliance(messages []*model.MessageForUnmarshal) (compliance []*common.NameValueChildren, err error) {
	var atsOkCount int
	var atsNotOkCount int
	var pciOkCount int
	var pciNotOkCount int
	var unknownCount int

	if len(messages) == 0 {
		return nil, nil
	}
	compliance = make([]*common.NameValueChildren, 2)

	for _, msg := range messages {
		if msg.Data == nil {
			unknownCount++
			continue
		}
		if msg.Progress == "done" {
			if msg.Data.Basic.IsATS {
				atsOkCount++
			} else {
				atsNotOkCount++
			}

			if msg.Data.Basic.IsPCI {
				pciOkCount++
			} else {
				pciNotOkCount++
			}
		}
	}
	ats := make([]*common.NameValue, 3)
	ats[0] = &common.NameValue{
		common.ATSAndPCIDSSSupport,
		atsOkCount,
	}
	ats[1] = &common.NameValue{
		common.ATSAndPCIDSSUnsupport,
		atsNotOkCount,
	}
	ats[2] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	compliance[0] = &common.NameValueChildren{
		common.ChartATS,
		ats,
	}

	pci := make([]*common.NameValue, 3)
	pci[0] = &common.NameValue{
		common.ATSAndPCIDSSSupport,
		pciOkCount,
	}
	pci[1] = &common.NameValue{
		common.ATSAndPCIDSSUnsupport,
		pciNotOkCount,
	}
	pci[2] = &common.NameValue{
		common.DashboardUnknown,
		unknownCount,
	}
	compliance[1] = &common.NameValueChildren{
		common.ChartPCIDSS,
		pci,
	}
	return compliance, nil
}

//区分出所有的证书信息

func findNameValue(names []*common.NameValue, name string) *common.NameValue {
	for _, v := range names {
		if v.Name == name {
			return v
		}
	}
	return nil
}
