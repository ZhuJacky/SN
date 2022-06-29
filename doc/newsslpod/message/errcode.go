// Package message provides ...
package message

const (
	// system error
	ErrSystem = -1
	// 0-100
	Success            = 0 // 操作成功 & operation success
	ErrLimitedRequest  = 1
	ErrIncorrectParams = 2

	// backend
	ErrInvalidEmail           = 1000
	ErrInvalidPhoneNo         = 1001
	ErrInvalidDomain          = 1002
	ErrLimitedAddDomain       = 1003
	ErrLimitedMonitorDomain   = 1004
	ErrRepetitionAdd          = 1005
	ErrInvalidServerType      = 1006
	ErrInvalidPort            = 1007
	ErrPunyCodeDomain         = 1008
	ErrInvalidIP              = 1009
	ErrInvalidTagName         = 1010
	ErrTooManyTag             = 1011
	ErrMoreThanAllow          = 1012
	ErrResolveDomainDNS       = 1013
	ErrInvalidDomainAccountId = 1014
	ErrInvalidId              = 1015
	ErrDomainClaimExists      = 1016
	ErrNotFoundDomainClaim    = 1017
	ErrClaimedDomain          = 1018
	ErrNotFoundRecordDNS      = 1019
	ErrNotFoundDomain         = 1020
	ErrUnverifiedDomainClaim  = 1021
	ErrNotBoundEmail          = 1022
	ErrNotBoundWechat         = 1023
	ErrNotBoundPhone          = 1024
	ErrInvalidSearchType      = 1025
	ErrNeedUpgradePlan        = 1026
	ErrInvalidProductId       = 1027
	ErrInvalidProductQuanlity = 1028
	ErrNotFoundAccountID      = 1029

	// checker
	ErrNotFoundChecker = 2000

	// notifier
	ErrUnsupportedNoticeType = 3000
)

var CodeDesc = map[int]string{
	ErrSystem:                 "系统错误",
	Success:                   "操作成功",
	ErrLimitedRequest:         "请求过快",
	ErrIncorrectParams:        "参数错误",
	ErrInvalidEmail:           "无效的邮箱地址",
	ErrInvalidPhoneNo:         "无效的手机号",
	ErrLimitedAddDomain:       "添加域名达到限制",
	ErrLimitedMonitorDomain:   "监控域名达到限制",
	ErrRepetitionAdd:          "重复添加",
	ErrInvalidServerType:      "无效的服务类型",
	ErrInvalidPort:            "无效的端口",
	ErrPunyCodeDomain:         "错误的域名",
	ErrInvalidIP:              "无效的IP",
	ErrInvalidTagName:         "无效的Tag名称",
	ErrTooManyTag:             "Tag过多",
	ErrMoreThanAllow:          "超过允许数量",
	ErrResolveDomainDNS:       "解析域名失败",
	ErrInvalidDomainAccountId: "无效的ID",
	ErrInvalidId:              "无效的ID",
	ErrDomainClaimExists:      "认领域名已存在",
	ErrNotFoundDomainClaim:    "未发现认领域名",
	ErrClaimedDomain:          "域名已验证",
	ErrNotFoundRecordDNS:      "未发现记录",
	ErrNotFoundDomain:         "域名不存在",
	ErrUnverifiedDomainClaim:  "未验证的域名",
	ErrNotBoundEmail:          "未绑定邮箱",
	ErrNotBoundWechat:         "未绑定微信",
	ErrNotBoundPhone:          "未绑定手机号",
	ErrInvalidSearchType:      "无效的搜索类型",
	ErrNeedUpgradePlan:        "需要升级订单",
	ErrInvalidProductId:       "无效的产品ID",
	ErrInvalidProductQuanlity: "无效的有效期",
	ErrNotFoundAccountID:      "未发现账户ID",

	// checker
	ErrNotFoundChecker: "未发现检测服务",

	// notifier
	ErrUnsupportedNoticeType: "不支持的通知类型",
}
