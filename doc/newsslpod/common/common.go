// Package common common
package common

// NameValue NameValue
type NameValue struct {
	Name  string
	Value int
}

// NameValueChildren NameValueChildren
type NameValueChildren struct {
	Name     string
	Children []*NameValue `json:",omitempty"`
}

// 等级
const (
	LevelUnknown uint8 = iota
	// APlus 等级A+
	LevelAPlus
	// A 等级A
	LevelA
	// ACut 等即A-
	LevelACut
	// B 等级B
	LevelB
	// C 等级C
	LevelC
	// D 等级D
	LevelD
	// E 等级E
	LevelE
	// F 等级F
	LevelF
	// T 等级T
	LevelT
)

// 安全评级描述
const (
	SecureLevelUnknown = "unknown"
	SecureLevelAPlus   = "A+"
	SecureLevelA       = "A"
	SecureLevelAMinus  = "A-"
	SecureLevelB       = "B"
	SecureLevelC       = "C"
	SecureLevelD       = "D"
	SecureLevelE       = "E"
	SecureLevelF       = "F"
	SecureLevelT       = "T"
	Unknown            = "未知"
)

// SecureLevelToInt SecureLevelToInt
func SecureLevelToInt(secureLevel string) uint8 {
	switch secureLevel {
	case SecureLevelUnknown:
		return LevelUnknown
	case SecureLevelAPlus:
		return LevelAPlus
	case SecureLevelA:
		return LevelA
	case SecureLevelAMinus:
		return LevelACut
	case SecureLevelB:
		return LevelB
	case SecureLevelC:
		return LevelC
	case SecureLevelD:
		return LevelD
	case SecureLevelE:
		return LevelE
	case SecureLevelF:
		return LevelF
	case SecureLevelT:
		return LevelT
	default:
		return 0
	}

}

// 证书状态
const (
	CannotConnect    = "连接异常"
	CertExpired      = "证书已过期"
	CertRevoke       = "证书已吊销"
	CertBlackList    = "证书黑名单"
	CertNotMatch     = "证书域名不匹配"
	CertUntrusted    = "证书不可信"
	CertUseWeakKey   = "证书密钥弱"
	CertExpiring7    = "证书即将过期, 少于7天"
	CertExpiring30   = "证书即将过期, 少于30天"
	CertTrust        = "正常"
	CertPartAbnormal = "部分异常"
)

// NewCertStatusText 新证书状态文案
var NewCertStatusText = map[string]string{
	"连接异常":          "连接异常",
	"证书已过期":         "已过期",
	"证书已吊销":         "已吊销",
	"证书黑名单":         "证书黑名单",
	"证书域名不匹配":       "域名不匹配",
	"证书不可信":         "不可信",
	"证书密钥弱":         "密钥弱",
	"证书即将过期, 少于7天":  "即将过期",
	"证书即将过期, 少于30天": "即将过期",
	"正常":            "正常",
	"部分异常":          "部分异常",
}

// 依证书严重性依次降低
const (
	UnknownStatusCode  = -1
	CannotConnectCode  = 0
	CertExpiredCode    = 1
	CertRevokeCode     = 2
	CertBlackListCode  = 3
	CertNotMatchCode   = 4
	CertUntrustedCode  = 5
	CertUseWeakKeyCode = 6
	CertValidityLt7    = 7
	CertValidityLt30   = 8
	CertTrustCode      = 9
)

// ChangeStatusToCode TODO
func ChangeStatusToCode(status string) int {
	switch status {
	case CannotConnect:
		return CannotConnectCode
	case CertExpired:
		return CertExpiredCode
	case CertRevoke:
		return CertRevokeCode
	case CertBlackList:
		return CertBlackListCode
	case CertNotMatch:
		return CertNotMatchCode
	case CertUntrusted:
		return CertUntrustedCode
	case CertUseWeakKey:
		return CertUseWeakKeyCode
	case CertExpiring7:
		return CertValidityLt7
	case CertExpiring30:
		return CertValidityLt30
	case CertTrust:
		return CertTrustCode
	}
	return UnknownStatusCode
}

// 合规状态
const (
	DashboardUnknown      = "未知"
	ATSAndPCIDSSSupport   = "合规"
	ATSAndPCIDSSUnsupport = "不合规"
	BugsAffect            = "受影响"
	BugsUnaffect          = "不受影响"
)

// 证书有效期
const (
	ValidityLt0   = "已过期"
	ValidityLt30  = "小于30天"
	ValidityLt60  = "小于60天"
	ValidityLt90  = "小于90天"
	ValidityGte90 = "大于90天"
)

// 漏洞信息
const (
	ChartATS                = "ATS"
	ChartPCIDSS             = "PCI DSS"
	ChartCetType            = "certType"
	ChartCertExpired        = "certExpired"
	BugOpenSSLPaddingOracle = "CVE-2016-2107" // OpenSSL Padding Oracle攻击
	BugDrown                = "CVE-2016-0800" // DROWN漏洞
	BugLogjam               = "CVE-2015-4000" // Logjam漏洞
	BugFreak                = "CVE-2015-0204" // FREAK漏洞
	BugPOODLE               = "CVE-2014-3566" // POODLE漏洞
	BugOpenSSLCCS           = "CVE-2014-0224" // OpenSSL CCS注入攻击
	BugHeartbleed           = "CVE-2014-0160" // 心血漏洞(Heartbleed)
	BugCRIME                = "CVE-2012-4929" // CRIME漏洞
)

// 搜索参数
const (
	SearchNone        = "none"
	SearchTags        = "tags"
	SearchSecureGrade = "grade"
	SearchBrand       = "brand"
	SearchCode        = "code"
	SearchHash        = "hash"
	SearchLimit       = "limit"
	SearchDomain      = "domain"
)

//
const (
	InvitationAdd = 2
)

// IsCorrectSecureCode 判断是否是正确的
func IsCorrectSecureCode(code string) bool {
	switch code {
	case Unknown, "Ap", SecureLevelAPlus, SecureLevelA, SecureLevelAMinus, SecureLevelB,
		SecureLevelC, SecureLevelD, SecureLevelE, SecureLevelF, SecureLevelT:
		return true
	default:
		return false
	}
}
