package db

import (
	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

var modelAccountDomain = &model.AccountDomain{}

//通过用户id获取关注域名集合（分页）
func GetDomainIdsLimitByAccountId(accountId int, start, limit int) (results []*model.AccountDomain, err error) {
	results = make([]*model.AccountDomain, 0)
	err = gormDB.Model(&model.AccountDomain{}).Where("account_id=?", accountId).Offset(start).Limit(limit).Order("id DESC").Find(&results).Error
	if err != nil {
		return nil, err
	}

	for _, v := range results {
		tags, err := GetTagsNameByDomainAccountId(v.Id)
		if err != nil {
			return nil, err
		}
		v.Tags = tags
	}

	return results, nil
}

//通过用户id获取所有用户关注的域名id
func GetDomainsIdByAccountId(accountId int) (ids []int, err error) {
	var results []*model.AccountDomain
	err = gormDB.Model(&model.AccountDomain{}).Where("account_id=?", accountId).Find(&results).Error
	if err != nil {
		return nil, err
	}
	for _, result := range results {
		ids = append(ids, result.DomainId)
	}
	return ids, nil
}

//得到用户关注域名的数量
func CountAttentionDomain(accountId int) (totoal int, err error) {
	var count int
	// err = gormDB.Model(&model.AccountDomain{}).Select("COUNT(DISTINCT domain)").
	// 	Joins("JOIN domain_results ON domain_results.id=domain_id").
	// 	Where("account_Id=?", accountId).Row().Scan(&count)
	err = gormDB.Model(&model.AccountDomain{}).Where("account_id=?", accountId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

//查询关注同一个域名的用户数量
func CountAttentionAccountAmount(domainId int) (number int, err error) {
	var count int
	err = gormDB.Model(&model.AccountDomain{}).Where("domain_id =? ", domainId).Count(&count).Error
	if err != nil {
		return 0, err
	}
	return count, nil
}

//插入用户域名关系表
func InsertAccountDomainRelation(relation *model.AccountDomain, tags []string) (err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	err = tx.Model(&model.AccountDomain{}).Create(relation).Error
	if err != nil {
		return
	}
	for _, v := range tags {
		err = tx.Create(&model.TagDomain{DomainAccountId: relation.Id, TagName: v}).Error
		if err != nil {
			return
		}
	}
	return
}

//用户已经关注过该域名
func UserHaveAttentionDomain(accountId, domainId int) (exist bool, err error) {
	result := new(model.AccountDomain)
	db := gormDB.Model(&model.AccountDomain{}).Where("account_id=? AND domain_id=?", accountId, domainId).First(&result)

	var recordNotFound bool

	if db.Error != nil {
		if !db.RecordNotFound() {
			return false, db.Error
		}
		recordNotFound = true
	}

	if recordNotFound {
		return false, nil
	}
	return true, nil
}

func DeleteRelation(accountId, domainId int) error {
	return gormDB.Model(&model.AccountDomain{}).Where("account_id=? AND domain_id=?", accountId, domainId).Delete(&model.AccountDomain{}).Error
}

func GetUsersByDomainId(domainId int) (results []*model.AccountDomain, err error) {
	err = gormDB.Model(&model.AccountDomain{}).Where("domain_id = ?", domainId).Find(&results).Error
	return
}

func UpdateAccountDomain(ad *model.AccountDomain) error {
	return gormDB.Save(ad).Error
}

// 更新是通知时间，如果勾选了通知选项
func ResetAccountDomainNoticedAt(domainId int) error {
	return gormDB.Model(&model.AccountDomain{}).Where("domain_id=? AND notice=true", domainId).Update("noticed_at", gorm.Expr("NULL")).Error
}

// 获取account_domains数据通过用户
func GetAccountDomainWithAccount(aid, domainAccountId int) (*model.AccountDomain, error) {
	ad := new(model.AccountDomain)
	err := gormDB.Where("account_id=? AND id=?", aid, domainAccountId).First(ad).Error
	if err != nil {
		return nil, err
	}

	rows, err := gormDB.Model(modelTagDomain).Select("tag_name").Where("domain_account_id=?", domainAccountId).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tags []string
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	ad.Tags = tags

	return ad, err
}

func GetAccountDomainWithAccountAndDomain(accountId, domainId int) (result *model.AccountDomain, err error) {
	result = new(model.AccountDomain)
	err = gormDB.Model(&model.AccountDomain{}).Where("account_id=? AND domain_id=?", accountId, domainId).First(result).Error
	return result, err
}

//设置需要通知
func SetNotice(accountId, domainId int, cancel bool) error {
	return gormDB.Model(&model.AccountDomain{}).Where("domain_id=? AND account_id=?", domainId, accountId).Update("notice", !cancel).Error
}

//获取通知状态
func GetNotice(accountId, domainId int) (notice bool, err error) {
	result := new(model.AccountDomain)
	err = gormDB.Model(&model.AccountDomain{}).Where("domain_id=? AND account_id=? ", domainId, accountId).First(result).Error
	notice = result.Notice
	return
}

//计算已经有多少个需要通知的域名
func CountNoticeDomainNumber(accountId int) (count int, err error) {
	err = gormDB.Model(&model.AccountDomain{}).Where("account_id=? AND notice=true", accountId).Count(&count).Error
	return
}

// 取消用户所有域名通知
func CancelNoticeByAccountId(accountId int) error {
	err := gormDB.Model(&model.AccountDomain{}).Where("account_id=?", accountId).Update("notice", false).Error
	return err
}

// 获取domain_account信息
func GetDomainAccountInfo(accountId, domainId int) (*model.AccountDomain, error) {
	accountDomain := new(model.AccountDomain)
	err := gormDB.Where("account_id=? AND domain_id=?", accountId, domainId).First(accountDomain).Error
	return accountDomain, err
}

// 获取监控站点数
func GetMonitorSiteCount() (int, error) {
	var total int
	err := gormDB.Model(modelAccountDomain).
		Count(&total).Error
	return total, err
}
