package model

import (
	"time"

	"mysslee_qcloud/common"
)

// DashboardResult 仪表盘表
type DashboardResult struct {
	Id        int       `gorm:"PRIMARY_KEY;AUTO_INCREMENT"` //id
	AccountId int       `gorm:"index;not null`              //账户id
	Result    string    `gorm:"type:longblob"`              //面板数据缓存
	UpdatedAt time.Time `gorm:"not null"`                   //上次更新时间
	CreatedAt time.Time `gorm:"default:now()"`              //创建时间
}

//饼图给数字
type DashboardShow struct {
	SecurityLevelPie         []*common.NameValue         //安全评估饼图
	CertBrandsPie            []*common.NameValue         //证书品牌饼图
	CertValidTimePie         []*common.NameValue         //证书有效期饼图
	CertTypePie              []*common.NameValue         //证书类型饼图
	SSLBugsLoopholeHistogram []*common.NameValueChildren //SSL漏洞分布柱状图
	ComplianceHistogram      []*common.NameValueChildren //合规性比例柱状图
}
