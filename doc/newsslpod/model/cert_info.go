// Package model provides ...
package model

import (
	"crypto/x509"
	"strings"
	"time"
)

// CertInfo 证书信息表
type CertInfo struct {
	Hash        string                  `gorm:"primary_key"`
	SN          string                  `gorm:"type:varchar(255)"` // 序列号
	CN          string                  `gorm:"type:varchar(255)"` // 通用名称
	SANs        string                  `gorm:"type:text"`         // 备用名称
	O           string                  `gorm:"type:varchar(255)"` // 组织
	OU          string                  `gorm:"type:varchar(255)"` // 组织单位
	Street      string                  `gorm:"type:varchar(255)"` // 街道
	City        string                  `gorm:"type:varchar(255)"` // 城市
	Province    string                  `gorm:"type:varchar(128)"` // 省份
	Country     string                  `gorm:"type:varchar(32)"`  // 国家
	KeyAlgo     string                  `gorm:"type:varchar(32)"`  // 私钥算法
	SignAlgo    x509.SignatureAlgorithm `gorm:"not null"`          // 签名算法
	CertType    string                  `gorm:"type:varchar(32)"`  // 证书类型
	BeginTime   time.Time               // 开始时间
	EndTime     time.Time               // 结束时间
	Issuer      string                  `gorm:"type:varchar(255)"` // 颁发者
	RawPEM      string                  `gorm:"type:text"`         // 原始数据
	Brand       string                  `gorm:"type:varchar(32)"`  // 证书品牌
	CreatedAt   time.Time               `gorm:"default:now()"`     // 创建时间
	TrustStatus string                  `gorm:"-"`                 // 信任状态
}

// DomainCert domain cert 关系表
type DomainCert struct {
	Id          int       `gorm:"primary_key;AUTO_INCREMENT"`
	DomainId    int       `gorm:"index;not null"`
	Hash        string    `gorm:"type:varchar(40);index;not null"`
	TrustStatus string    `gorm:"type:varchar(32)"` // 信任状态
	NoticedAt   time.Time `gorm:"default:null"`
	CreatedAt   time.Time `gorm:"default:now()"`
}

type CertInfoShow struct {
	Hash        string
	CN          string
	SANs        string
	KeyAlgo     string
	Issuer      string
	BeginTime   time.Time
	EndTime     time.Time
	Days        int
	Brand       string
	TrustStatus string
	CertType    string
}

func CertInfoForShow(infos []*CertInfo) []*CertInfoShow {
	shows := make([]*CertInfoShow, len(infos))
	for i, v := range infos {
		shows[i] = &CertInfoShow{
			Hash:        v.Hash,
			CN:          v.CN,
			SANs:        v.SANs,
			KeyAlgo:     v.KeyAlgo,
			Issuer:      v.Issuer,
			BeginTime:   v.BeginTime,
			EndTime:     v.EndTime,
			Brand:       v.Brand,
			Days:        int(v.EndTime.UTC().Sub(time.Now().UTC()).Hours() / 24),
			TrustStatus: strings.Split(v.TrustStatus, ",")[0],
			CertType:    v.CertType,
		}
	}
	return shows
}
