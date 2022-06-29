// Package model provides ...
package model

import (
	"encoding/json"
	"time"
)

const (
	SSLPodV1 = "sslpod_v1"
	SSLPodV2 = "sslpod_v2"
	SSLPodV3 = "sslpod_v3"
)

const (
	MaxMonitorLimitName = "limit_monitor"
	MaxAddLimitName     = "limit_maxadd"
	EmailLimitName      = "limit_email"
	WechatLimitName     = "limit_wechat"
	PhoneLimitName      = "limit_phone"
)

// Product table
type Product struct {
	Id          int       `gorm:"primary_key"`
	Pid         int       `gorm:"unique_index;not null"`      // 计费侧定义
	Name        string    `gorm:"type:varchar(64);not null"`  // 名称
	Description string    `gorm:"type:varchar(255);not null"` // 描述
	TimeUnit    string    `gorm:"not null"`                   // 时间单位 y,m,d,h
	UnitPrice   float64   `gorm:"not null"`                   // 价格
	Content     []byte    `gorm:"not null" json:"-"`          // 产品详细
	UpdatedAt   time.Time `gorm:"default:'1991-09-10 00:00:00'"`
	CreatedAt   time.Time `gorm:"default:now()"`

	// 用来序列化和凡序列化
	ContentRawMessage json.RawMessage `gorm:"-" json:"content"`
}

// ProductContent for product.Content
type ProductContent struct {
	MaxAllowMonitoringCount int // 该套餐允许最多监控域名
	MaxAllowAddDomainCount  int // 该套餐允许最多添加域名
	MaxAllowEmailWarnCount  int // 该套餐允许最多邮件告警
	MaxAllowWechatWarnCount int // 该套餐允许最多微信告警
	MaxAllowPhoneWarnCount  int // 该套餐允许最多微信告警
}

// PlanInfo account current plan
type PlanInfo struct {
	ProductContent
	Name      string
	ExpiredAt string
	Pid       int
}
