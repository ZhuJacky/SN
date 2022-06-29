package model

import "time"

// TagDomain 标签域名表
type TagDomain struct {
	DomainAccountId int       `gorm:"primary_key"`
	TagName         string    `gorm:"type:varchar(64);primary_key"`
	CreatedAt       time.Time `gorm:"default:now()"`
}
