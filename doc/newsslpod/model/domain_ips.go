package model

// DomainIps 多地域检测结果表
type DomainIps struct {
	DomainID     int    `gorm:"type:int;primary_key;not null"`
	Uin          string `gorm:"type:varchar(64);primary_key;not null"`
	IpPorts      string `gorm:"type:text"`
	IsAutoDetect bool   `gorm:"type:tinyint;default 1"`
}
