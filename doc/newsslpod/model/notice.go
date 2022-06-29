// Package model provides ...
package model

import "time"

// 通知方式，限制8种
const (
	NoticeWechat        = 1 << iota // 1 微信
	NoticePhoneSMS      = 1 << iota // 2 短信
	NoticeEmail         = 1 << iota // 4 邮件
	NoticeStationLetter = 1 << iota // 8 站内信
	NoticeNone          = 0         // 关闭通知
	NoticeAll           = NoticeWechat | NoticePhoneSMS | NoticeEmail | NoticeStationLetter
)

// 通知类型
const (
	NoticeTypeCertStatus = 1 + iota // 证书状态不正常
	NoticeTypeCSPStatus             // 不安全外链告警
)

func NoticeString(typ int) string {
	switch typ {
	case NoticeWechat:
		return "wechat"
	case NoticePhoneSMS:
		return "phoneSMS"
	case NoticeEmail:
		return "email"
	case NoticeStationLetter:
		return "stationLetter"
	}
	return "unknown"
}

func NoticeInt(str string) int {
	switch str {
	case "wechat":
		return NoticeWechat
	case "phoneSMS":
		return NoticePhoneSMS
	case "email":
		return NoticeEmail
	case "stationLetter":
		return NoticeStationLetter
	}
	return -1
}

func IsSupportedNoticeType(typ int) bool {
	if typ&NoticeStationLetter == NoticeStationLetter {
		return false
	}
	return typ >= NoticeNone &&
		typ <= (NoticeWechat|NoticePhoneSMS|NoticeEmail)
}

// NoticeMsg 告警通知信息表
type NoticeMsg struct {
	Id            int       `gorm:"primary_key;AUTO_INCREMENT"`
	Uin           string    `gorm:"type:varchar(64);not null;index"`
	Type          int       `gorm:"not null"`                       // 通知类型
	Msg           string    `gorm:"type:varchar(255);not null"`     // 通知内容, map
	Language      string    `gorm:"type:varchar(16);not null"`      // 语言
	NoticedResult string    `gorm:"type:varchar(128);default:null"` // 通知结果
	NoticeAlready int       `gorm:"default:null"`                   // 已通知
	NoticedAt     time.Time `gorm:"default:null";index`             // 通知时间
	CreatedAt     time.Time `gorm:"default:now()"`                  // 创建时间
}

// NoticeInfo 通知开关表
type NoticeInfo struct {
	Id         int       `gorm:"primary_key;AUTO_INCREMENT"`
	Uin        string    `gorm:"type:varchar(64);not null;unique_index"`
	NoticeType int       `gorm:"index;not null"`
	UpdatedAt  time.Time `gorm:"default:'1991-09-10 00:00:00'"`
	CreatedAt  time.Time `gorm:"default:now()"`
}
