// Package model provides ...
package model

import "time"

// 账号状态
const (
	StatusUnregistered = iota // 0 未注册
	StatusUnactivated         // 1 未激活
	StatusActivated           // 2 已激活
	StatusForbidden    = -1   // -1 拉黑
)

// Account 用户表
type Account struct {
	Id        int       `gorm:"primary_key;AUTO_INCREMENT"`
	Uin       string    `gorm:"type:varchar(64);unique_index"`
	Name      string    `gorm:"type:varchar(64)"`
	Status    int       `gorm:"not null"`
	Aggregate int       `gorm:"default:0"` // 是否需要聚合，0 不聚合
	CreatedAt time.Time `gorm:"default:now()"`
}

// AccountShow 前端显示结构
type AccountShow struct {
	Name      string
	Status    int
	CreatedAt time.Time
	Plan      *PlanInfo // 计划（套餐）ID
}

type LimitInfo struct {
	Type  string // 额度类型
	Total int    // 额度
	Sent  int    // 已使用
}

func AccountForShow(a *Account) *AccountShow {
	as := new(AccountShow)
	as.Name = a.Name
	as.Status = a.Status
	as.CreatedAt = a.CreatedAt
	return as
}
