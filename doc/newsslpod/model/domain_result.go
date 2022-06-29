package model

import (
	"encoding/json"
	"math"
	"time"

	"mysslee_qcloud/common"
)

const (
	DomainFlagBindIP = 1 << iota
)

// 以 domain,port,ip,servertype 做联合查询
type DomainResult struct {
	Id                    int       `gorm:"PRIMARY_KEY;AUTO_INCREMENT"`                                                //id
	Domain                string    `gorm:"type:varchar(255);index;not null"`                                          //域名
	IP                    string    `gorm:"type:varchar(64);index;column:ip"`                                          //ip
	PunyCodeDomain        string    `gorm:"type:varchar(255);column:punycode_domain;not null"`                         //编码后的域名
	Port                  string    `gorm:"type:varchar(64);index;not null"`                                           //端口
	ServerType            int       `gorm:"column:server_type;default:0;index"`                                        //检测的服务器类型
	LastFastDetectionTime time.Time `gorm:"column:last_fast_detection_time;index"`                                     //上次快速检测时间
	LastFullDetectionTime time.Time `gorm:"column:last_full_detection_time;index"`                                     //上次完全检测时间
	FullDetectionResult   []byte    `gorm:"type:longblob"`                                                             //上次完全检测的时间
	TrustStatus           string    `gorm:"type:varchar(64);column:trust_status"`                                      //证书信任状态
	PrevStatus            string    `gorm:"type:varchar(64);column:prev_status"`                                       //上一次状态
	LoseEfficacy          bool      `gorm:"column:lose_efficacy;index"`                                                //失效标记
	Brand                 string    `gorm:"type:varchar(32);column:brand"`                                             //证书品牌
	Grade                 string    `gorm:"type:varchar(8);column:grade"`                                              //评级
	CreatedAt             time.Time `gorm:"column:created_at;index"`                                                   //创建时间
	DomainStatus          string    `gorm:"column:domain_status;default: '{\"brands\":[\"未知\"],\"status\":17825792}'"` //整体状态
	ResultHash            string    `gorm:"type:varchar(40)"`                                                          // 结果hash，用来判断是否需要聚合
	DomainFlag            int       `grom:"not null;default:0;index"`

	// need not save
	AccountDomainId int             `gorm:"-"`
	Notice          bool            `gorm:"-"`
	Tags            []string        `gorm:"-"`
	Children        []*DomainResult `gorm:"-"`
}

type DomainStatus struct {
	Brands []string `json:"brands"` //证书品牌
	Status int64    `json:"status"` //状态
}

const (
	ATSCompliance    int64 = 0x03 << ShiftForATS         //两bit位表示ATS状态
	PCIDSSCompliance       = 0x03 << ShiftForPCIDSS      //两bit表示PCIDSS状态
	CVE20162107            = 0x03 << ShiftForCVE20162107 //两bit表示CVE-2016-2107状态
	CVE20160800            = 0x03 << ShiftForCVE20160800 //两bit表示CVE-2016-0800状态
	CVE20154000            = 0x03 << ShiftForCVE20154000 //两bit表示CVE-2015-4000状态
	CVE20150204            = 0x03 << ShiftForCVE20150204 //两bit表示CVE-2015-0240状态
	CVE20143566            = 0x03 << ShiftForCVE20143566 //两bit表示CVE-2014-3566状态
	CVE20140224            = 0x03 << ShiftForCVE20140224 //两bit表示CVE-2014-0224状态
	CVE20140160            = 0x03 << ShiftForCVE20140160 //两bit表示CVE-2014-0160状态
	CVE20124929            = 0x03 << ShiftForCVE20124929 //两bit表示CVE-2012-4929状态
	CertType               = 0x0f << ShiftForCertType    //三bit表示证书类别
	CertExpired            = 0x7f << ShiftForCertExpired //七bit表示证书类别
)

func GetBitAndNumber(shift uint) int64 {
	switch shift {
	case ShiftForATS:
		return ATSCompliance
	case ShiftForPCIDSS:
		return PCIDSSCompliance
	case ShiftForCVE20162107:
		return CVE20162107
	case ShiftForCVE20160800:
		return CVE20160800
	case ShiftForCVE20154000:
		return CVE20154000
	case ShiftForCVE20150204:
		return CVE20150204
	case ShiftForCVE20143566:
		return CVE20143566
	case ShiftForCVE20140224:
		return CVE20140224
	case ShiftForCVE20140160:
		return CVE20140160
	case ShiftForCVE20124929:
		return CVE20124929
	case ShiftForCertType:
		return CertType
	case ShiftForCertExpired:
		return CertExpired
	default:
		return ATSCompliance
	}

}

//需要移位的数量
//TODO 之后使用三位bit作为标志位
const (
	ShiftForATS         uint = 0
	ShiftForPCIDSS      uint = 2
	ShiftForCVE20162107 uint = 4
	ShiftForCVE20160800 uint = 6
	ShiftForCVE20154000 uint = 8
	ShiftForCVE20150204 uint = 10
	ShiftForCVE20143566 uint = 12
	ShiftForCVE20140224 uint = 14
	ShiftForCVE20140160 uint = 16
	ShiftForCVE20124929 uint = 18
	ShiftForCertType    uint = 20
	ShiftForCertExpired uint = 24
)

//判断是否是正确的
func IsCurrentShift(shift uint) bool {
	switch shift {
	case ShiftForATS, ShiftForPCIDSS, ShiftForCVE20162107, ShiftForCVE20160800,
		ShiftForCVE20154000, ShiftForCVE20150204, ShiftForCVE20143566, ShiftForCVE20140224,
		ShiftForCVE20140160, ShiftForCVE20124929, ShiftForCertType, ShiftForCertExpired:
		return true
	default:
		return false
	}
}

const (
	//TODO 之后使用3位0x01，0x02，0x04，
	SupportUnknown = 0x00
	Support        = 0x01
	NotSupport     = 0x02
)

const (
	CertTypeUnknown = 0x01
	CertTypeDV      = 0x02
	CertTypeOV      = 0x04
	CertTypeEV      = 0x08
)

const (
	CertExpireUnknown = 0x01
	CertExpireLt0     = 0x02
	CertExpireLt30    = 0x04
	CertExpireLt60    = 0x08
	CertExpireLt90    = 0x10
	CertExpireGte90   = 0x20
)

//计算
func CalculateStatusForUnknown() *DomainStatus {
	var domainStatus = &DomainStatus{}
	domainStatus.Brands = []string{"未知"}
	domainStatus.Status = CertTypeUnknown<<ShiftForCertType | CertExpireUnknown<<ShiftForCertExpired

	return domainStatus
}

//计算状态
func CalculateStatus(report *Report) *DomainStatus {
	domainStatus := &DomainStatus{}
	//先获取证书品牌
	domainStatus.Brands = getBrandName(report.Certs)
	atsStatus := calATSCompliance(report.Basic.IsATS)
	pcidssStatus := calPCIDSSCompliance(report.Basic.IsPCI)
	cve20162107 := calCVE(report.Bugs.PaddingOracle.Support, ShiftForCVE20162107)
	cve20160800 := calCVE(report.Bugs.Drown.Support, ShiftForCVE20160800)
	cve20154000 := calCVE(report.Bugs.Logjam.Support, ShiftForCVE20154000)
	cve20150204 := calCVE(report.Bugs.FREAK.Support, ShiftForCVE20150204)
	cve20143566 := calCVE(report.Bugs.POODLE.Support, ShiftForCVE20143566)
	cve20140224 := calCVE(report.Bugs.CCS.Support, ShiftForCVE20140224)
	cve20140160 := calCVE(report.Bugs.Heartbleed.Support, ShiftForCVE20140160)
	cve20124929 := calCVE(report.Bugs.CRIME.Support, ShiftForCVE20124929)
	certTypeStatus := calCertType(report)
	certExpiredStatus := calCertExpired(report)
	domainStatus.Status = atsStatus | pcidssStatus | cve20162107 | cve20160800 | cve20154000 | cve20150204 | cve20143566 | cve20140224 |
		cve20140160 | cve20124929 | certTypeStatus | certExpiredStatus

	return domainStatus
}

// CalculateStatusAll 计算全部状态
func CalculateStatusAll(status string, report *Report) *DomainStatus {
	domainStatus := &DomainStatus{}
	_ = json.Unmarshal([]byte(status), domainStatus) // ignore error

	atsStatus := calATSCompliance(report.Basic.IsATS)
	pcidssStatus := calPCIDSSCompliance(report.Basic.IsPCI)
	cve20162107 := calCVE(report.Bugs.PaddingOracle.Support, ShiftForCVE20162107)
	cve20160800 := calCVE(report.Bugs.Drown.Support, ShiftForCVE20160800)
	cve20154000 := calCVE(report.Bugs.Logjam.Support, ShiftForCVE20154000)
	cve20150204 := calCVE(report.Bugs.FREAK.Support, ShiftForCVE20150204)
	cve20143566 := calCVE(report.Bugs.POODLE.Support, ShiftForCVE20143566)
	cve20140224 := calCVE(report.Bugs.CCS.Support, ShiftForCVE20140224)
	cve20140160 := calCVE(report.Bugs.Heartbleed.Support, ShiftForCVE20140160)
	cve20124929 := calCVE(report.Bugs.CRIME.Support, ShiftForCVE20124929)

	// clear
	domainStatus.Status &= 255 << 20

	domainStatus.Status |= atsStatus | pcidssStatus | cve20162107 |
		cve20160800 | cve20154000 | cve20150204 |
		cve20143566 | cve20140224 | cve20140160 | cve20124929
	return domainStatus
}

// CalculateStatusCert 只计算证书状态
func CalculateStatusCert(status string, infos []*CertInfo) *DomainStatus {
	domainStatus := &DomainStatus{}
	_ = json.Unmarshal([]byte(status), domainStatus) // ignore error

	var (
		brands  []string
		types   int64
		expires int64
	)
	for _, info := range infos {
		// brands
		brands = append(brands, info.Brand)
		// cert types
		switch info.CertType {
		case "NoAudit":
			types |= CertTypeUnknown << ShiftForCertType
		case "DV":
			types |= CertTypeDV << ShiftForCertType
		case "OV":
			types |= CertTypeOV << ShiftForCertType
		case "EV":
			types |= CertTypeEV << ShiftForCertType
		}
		// expire status
		days := math.Ceil(info.EndTime.UTC().Sub(time.Now().UTC()).Hours() / 24)
		if days <= 0 {
			expires |= CertExpireLt0 << ShiftForCertExpired
		} else if days < 30 {
			expires |= CertExpireLt30 << ShiftForCertExpired
		} else if days < 60 {
			expires |= CertExpireLt60 << ShiftForCertExpired
		} else if days < 90 {
			expires |= CertExpireLt90 << ShiftForCertExpired
		} else {
			expires |= CertExpireGte90 << ShiftForCertExpired
		}
	}

	domainStatus.Brands = brands
	domainStatus.Status &= 2<<20 - 1
	domainStatus.Status |= types | expires
	return domainStatus
}

func getBrandName(certInfo *MultipleCertificateInfo) []string {
	var brands = make([]string, 0)
	var brandMap = make(map[string]bool)
	var sniBrands = make([]string, 0)   //通过SNI获取的证书
	var nosniBrands = make([]string, 0) //通过非SNI获取的证书

	for _, cert := range certInfo.RSAs {
		if cert.LeafCertInfo.IsSNI {
			sniBrands = append(sniBrands, cert.LeafCertInfo.BrandName)
		} else {
			nosniBrands = append(nosniBrands, cert.LeafCertInfo.BrandName)
		}
	}

	for _, cert := range certInfo.ECCs {
		if cert.LeafCertInfo.IsSNI {
			sniBrands = append(sniBrands, cert.LeafCertInfo.BrandName)
		} else {
			nosniBrands = append(nosniBrands, cert.LeafCertInfo.BrandName)
		}
	}

	if len(sniBrands) != 0 {
		for _, brand := range sniBrands {
			brandMap[brand] = true
		}
	} else {
		for _, brand := range nosniBrands {
			brandMap[brand] = true
		}
	}

	for key, _ := range brandMap {
		brands = append(brands, key)
	}
	return brands
}

//计算ATS状态
func calATSCompliance(ats bool) int64 {
	if ats {
		return Support << ShiftForATS
	} else {
		return NotSupport << ShiftForATS
	}
}

//计算PCIDSS
func calPCIDSSCompliance(pcidss bool) int64 {
	if pcidss {
		return Support << ShiftForPCIDSS
	} else {
		return NotSupport << ShiftForPCIDSS
	}
}

//计算漏洞信息
func calCVE(result int, shiftForCVE uint) int64 {
	if result == BugUnknown {
		return SupportUnknown << shiftForCVE
	} else if result == BugUnsupported {
		return NotSupport << shiftForCVE
	} else {
		return Support << shiftForCVE
	}
}

//计算证书类型
func calCertType(report *Report) int64 {
	sni, nosni := DistinguishCerts(report)
	if len(sni) > 0 {
		return getCertTypeResult(sni)
	} else {
		return getCertTypeResult(nosni)
	}
}

func getCertTypeResult(certs []*LeafCertInfo) int64 {
	var result int64 = 0
	for _, cert := range certs {
		switch cert.CertType {
		case "NoAudit":
			result |= CertTypeUnknown << ShiftForCertType
		case "DV":
			result |= CertTypeDV << ShiftForCertType
		case "OV":
			result |= CertTypeOV << ShiftForCertType
		case "EV":
			result |= CertTypeEV << ShiftForCertType
		}

	}
	return result
}

//计算证书有效期
func calCertExpired(report *Report) int64 {
	sni, nosni := DistinguishCerts(report)
	if len(sni) > 0 {
		return getCertExpiredResult(sni)
	} else {
		return getCertExpiredResult(nosni)
	}
}

func getCertExpiredResult(certs []*LeafCertInfo) int64 {
	var result int64 = 0
	for _, cert := range certs {
		scope := GetCertValidaty(cert)
		switch scope {
		case common.ValidityLt0:
			result |= CertExpireLt0 << ShiftForCertExpired
		case common.ValidityLt30:
			result |= CertExpireLt30 << ShiftForCertExpired
		case common.ValidityLt60:
			result |= CertExpireLt60 << ShiftForCertExpired
		case common.ValidityLt90:
			result |= CertExpireLt90 << ShiftForCertExpired
		case common.ValidityGte90:
			result |= CertExpireGte90 << ShiftForCertExpired
		default:
			result |= CertExpireUnknown << ShiftForCertExpired
		}
	}
	return result
}

func VerifyCode(status int64, code int64, shift uint) bool { //这里需要对证书有效，证书类型和其他的进行区分
	s := GetBitAndNumber(shift)
	result := (status & s) >> shift
	switch shift {
	case ShiftForCertExpired, ShiftForCertType:
		if result&code > 0 {
			return true
		}
		return false
	default:
		if result == code {
			return true
		}
		return false
	}

}

func GetShiftFromItemType(itemType string) (shift uint) {
	switch itemType {
	case common.ChartATS:
		return ShiftForATS
	case common.ChartPCIDSS:
		return ShiftForPCIDSS
	case common.BugOpenSSLPaddingOracle:
		return ShiftForCVE20162107
	case common.BugDrown:
		return ShiftForCVE20160800
	case common.BugLogjam:
		return ShiftForCVE20154000
	case common.BugFreak:
		return ShiftForCVE20150204
	case common.BugPOODLE:
		return ShiftForCVE20143566
	case common.BugOpenSSLCCS:
		return ShiftForCVE20140224
	case common.BugHeartbleed:
		return ShiftForCVE20140160
	case common.BugCRIME:
		return ShiftForCVE20124929
	case common.ChartCetType:
		return ShiftForCertType
	case common.ChartCertExpired:
		return ShiftForCertExpired
	}
	return 0
}

func GetExpiredCode(s string) int64 {
	switch s {
	case common.ValidityLt0:
		return CertExpireLt0
	case common.ValidityLt30:
		return CertExpireLt30
	case common.ValidityLt60:
		return CertExpireLt60
	case common.ValidityLt90:
		return CertExpireLt90
	case common.ValidityGte90:
		return CertExpireGte90
	}
	return CertExpireUnknown
}

func GetCertTypeCode(s string) int64 {
	switch s {
	case "DV":
		return CertTypeDV
	case "OV":
		return CertTypeOV
	case "EV":
		return CertTypeEV
	}
	return CertTypeUnknown
}

func GetSupportTypeCode(s string) int64 {
	switch s {
	case common.DashboardUnknown:
		return SupportUnknown
	case common.BugsAffect, common.ATSAndPCIDSSSupport:
		return Support
	case common.BugsUnaffect, common.ATSAndPCIDSSUnsupport:
		return NotSupport
	default:
		return SupportUnknown
	}
}

func GetCodeFromShift(shift uint, codeType string) int64 {
	switch shift {
	case ShiftForATS, ShiftForPCIDSS, ShiftForCVE20162107, ShiftForCVE20160800,
		ShiftForCVE20154000, ShiftForCVE20150204, ShiftForCVE20143566, ShiftForCVE20140224,
		ShiftForCVE20140160, ShiftForCVE20124929:
		return GetSupportTypeCode(codeType)
	case ShiftForCertType:
		return GetCertTypeCode(codeType)
	case ShiftForCertExpired:
		return GetExpiredCode(codeType)
	}
	return 0
}
