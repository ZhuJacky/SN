// Package model provides ...
package model

import (
	"encoding/json"
	"time"
)

const (
	TIME_FORMAT = "2006-01-02 15:04:05"
	TIME_ZERO   = "0000-00-00 00:00:00"
)

var (
	// 应对 MySQL timestamp存储时间问题
	TimeZeroAt = time.Date(1991, 9, 10, 0, 0, 0, 0, time.UTC)
)

// Resource 订单表
type Resource struct {
	ResourceId        string    `gorm:"primary_key" json:"resourceId"`                // 资源ID
	Uin               string    `gorm:"type:varchar(64);index;not null" json:"uin"`   // 资源拥有者
	AppId             int       `gorm:"type:varchar(64);index;not null" json:"appId"` // 云平台应用ID，与uin一一对应
	ProjectId         int       `gorm:"default:0" json:"projectId`                    // 云平台项目ID
	RenewFlag         int       `gorm:"default:0" json:"renewFlag"`                   // 自动续费标记 0未设置 1续费
	Region            int       `gorm:"default:0" json:"region"`                      // 地域ID
	ZoneId            int       `gorm:"not null" json:"zoneId"`                       // 区域ID
	Status            int       `gorm:"not null" json:"status"`                       // 资源状态，1正常 2隔离 3销毁
	PayMode           int       `gorm:"not null" json:"payMode"`                      // 付费模式 0 按需付费 1预付费
	IsolatedTimestamp time.Time `json:"isolatedTimestamp"`                            // 资源隔离时间
	CreateTime        time.Time `gorm:"default:now()" json:"createTime"`              // 创建时间
	ExpireTime        time.Time `gorm:"index" json:"expireTime"`                      // 资源到期时间
	GoodsDetail       []byte    `gorm:"not null" json:"goodsDetail"`
}

func (r *Resource) GetGoodsDetail() (*GoodsDetail, error) {
	detail := new(GoodsDetail)
	err := json.Unmarshal(r.GoodsDetail, detail)
	return detail, err
}

type GoodsDetail struct {
	Pid           int    `json:"pid"`
	TimeSpan      int    `json:"timeSpan"`
	TimeUnit      string `json:"timeUnit"`
	GoodsNum      int    `json:"goodsNum"`
	AutoRenewFlag int    `json:"autoRenewFlag"`
	SSLPodV1      int    `json:"sslpod_v1,omitempty"`
	SSLPodV2      int    `json:"sslpod_v2,omitempty"`
	SSLPodV3      int    `json:"sslpod_v3,omitempty"`
}

const (
	OrderStatusValid   = "valid"
	OrderStatusRenew   = "renew"
	OrderStatusRefund  = "refund"
	OrderStatusDestroy = "destroy"
)

// Order 订单表
type Order struct {
	Id         int       `gorm:"primary_key;auto_increment"`
	TranId     string    `gorm:"type:varchar(32);primary_key"`
	ResourceId string    `gorm:"type:varchar(32);index;not null"`
	Status     string    `gorm:"type:varchar(16);default:'valid'"`
	UpdatedAt  time.Time `gorm:"default:'1991-09-10 00:00:00'"`
	CreatedAt  time.Time `gorm:"default:now()"`
}
