// Package model provides ...
package model

type LimitConsume struct {
	Uin  string `gorm:"type:varchar(64);primary_key"`
	Date string `gorm:"type:varchar(20);primary_key"` // 2006-01
	Cost int    `gorm:"not null;default:0"`
}
