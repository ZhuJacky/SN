package model

import (
	"strings"
	"time"

	"mysslee_qcloud/common"
)

type AccountDomain struct {
	Id        int       `gorm:"PRIMARY_KEY;AUTO_INCREMENT"` //记录id
	AccountId int       `gorm:"index;not null"`             //账户id
	DomainId  int       `gorm:"index;not null"`             //域名id
	Notice    bool      `gorm:"default:false"`              //是否需要通知
	CreatedAt time.Time `gorm:"default:now()"`              //创建时间
	NoticedAt time.Time `gorm:"default:null"`

	// need to save
	Tags []string `gorm:"-"`
}

//站点信息
type SiteInfo struct {
	Id              int
	Domain          string
	Ip              string
	AutoIP          bool
	Port            string
	ServerType      int
	Brand           string
	Status          string
	Grade           string
	GradeCode       uint8
	Notice          bool
	AccountDomainId int
	Tags            []string
	Children        []*SiteInfo `json:",omitempty"`
}

//把DomainResult信息转换成SiteInfo
func ChangeDomainResultToSiteInfo(result *DomainResult) *SiteInfo {
	info := &SiteInfo{}
	info.Id = result.Id
	info.Domain = result.Domain
	info.Ip = result.IP
	info.Port = result.Port
	if result.Brand != "" && result.Brand != "unknown" {
		info.Brand = result.Brand
	} else {
		info.Brand = ""
	}

	info.ServerType = result.ServerType
	info.Status = strings.Split(result.TrustStatus, ",")[0]
	info.GradeCode = common.SecureLevelToInt(result.Grade)
	if result.Grade == common.SecureLevelUnknown {
		info.Grade = ""
	} else {
		info.Grade = result.Grade
	}

	info.Notice = result.Notice
	info.AccountDomainId = result.AccountDomainId
	info.Tags = result.Tags
	info.AutoIP = result.DomainFlag&DomainFlagBindIP == 0
	return info
}
