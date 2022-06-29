package model

import "time"

const (
	DomainClaimStatusNone     = iota + 1 // 未认证
	DomainClaimStatusFailed              // 认证失败
	DomainClaimStatusVerified            // 认证成功
)

// DomainClaim 域名认领表
type DomainClaim struct {
	Id        int       `gorm:"PRIMARY_KEY;AUTO_INCREMENT"` //域名id
	AccountId int       `gorm:"index;not null"`             //用户id
	Domain    string    `gorm:"type:varchar(255);not null"` //认领的域名
	Status    int       `gorm:"not null"`                   //认领域名状态
	TXTRecord string    `gorm:"type:varchar(64);not null"`  //TXT记录值
	CreatedAt time.Time `gorm:"default:now()"`              //创建时间
}

type DomainClaimShow struct {
	Id     int    `json:"id"`
	Domain string `json:"domain"`
	Status int    `json:"status"`
	Hidden bool   `json:"hidden"`
}
