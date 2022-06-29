package db

import (
	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

var modelAccountDomain = &model.AccountDomain{}

// 更新是通知时间，如果勾选了通知选项
func ResetAccountDomainNoticedAt(domainId int) error {
	return gormDB.Model(&model.AccountDomain{}).
		Where("domain_id=? AND notice=true", domainId).
		Update("noticed_at", gorm.Expr("NULL")).Error
}

//设置需要通知
func SetNotice(accountId, domainId int, cancel bool) error {
	return gormDB.Model(&model.AccountDomain{}).
		Where("domain_id=? AND account_id=?", domainId, accountId).Update("notice", !cancel).Error
}

//获取通知状态
func GetNotice(accountId, domainId int) (notice bool, err error) {
	result := new(model.AccountDomain)
	err = gormDB.Model(&model.AccountDomain{}).
		Where("domain_id=? AND account_id=? ", domainId, accountId).First(result).Error
	notice = result.Notice
	return
}

// 获取domain_account信息
func GetDomainAccountInfo(accountId, domainId int) (*model.AccountDomain, error) {
	accountDomain := new(model.AccountDomain)
	err := gormDB.Where("account_id=? AND domain_id=?", accountId, domainId).First(accountDomain).Error
	return accountDomain, err
}

// 获取监控该域名的用户
func GetUsersWatchDomain(domainId int) ([]*model.AccountDomain, error) {
	var results []*model.AccountDomain
	err = gormDB.Model(modelAccountDomain).
		Where("domain_id=? AND notice=true", domainId).Find(&results).Error
	return results, err
}

// 更新关系表
func UpdateAccountDomain(id int, fields map[string]interface{}) error {
	return gormDB.Model(modelAccountDomain).
		Where("id=?", id).Updates(fields).Error
}
