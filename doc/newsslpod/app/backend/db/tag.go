// Package db provides ...
package db

import (
	"sort"

	"mysslee_qcloud/model"

	"github.com/jinzhu/gorm"
)

var modelTagDomain = &model.TagDomain{}

//通过域名账号id获取对应的标签信息
func GetTagsNameByDomainAccountId(id int) (tags []string, err error) {
	rows, err := gormDB.Model(modelTagDomain).Select("tag_name").Where("domain_account_id=?", id).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	return tags, nil
}

func GetAccountTags(aid int) ([]string, error) {
	tags := []string{}
	rows, err := gormDB.Model(modelTagDomain).Select("DISTINCT(tag_name)").
		Joins("JOIN account_domains ON account_domains.id=domain_account_id").
		Where("account_id=?", aid).Rows()
	defer rows.Close()
	for rows.Next() {
		var tag string
		err = rows.Scan(&tag)
		if err != nil {
			return nil, err
		}
		tags = append(tags, tag)
	}
	sort.Strings(tags)
	return tags, err
}

//搜索标签
func SearchTags(username string, tag string) (tags []string, err error) {
	err = gormDB.Model(modelTagDomain).Select("DISTINCT(tag_name)").
		Joins("JOIN account_domains ON account_domains.id=domain_account_id").
		Joins("JOIN accounts ON account_id=accounts.id").
		Where("tag_name like ? AND (accounts.email=? OR accounts.phone=?)", tag, username, username).Scan(&tags).Error
	return tags, err
}

//计算用户选择的tag的域名数量
func CountAccountTags(username, tag string) (count int, err error) {
	rows, err := gormDB.Model(modelTagDomain).Select("DISTINCT tag_name").
		Joins("JOIN  account_domains on account_domains.id=domain_account_id ").
		Joins("JOIN  accounts on account_id=accounts.id ").
		Where("(accounts.email=? OR accounts.phone=?) AND tag_domains.tag_name like ?", username, username, tag).Rows()
	if err != nil {
		return 0, err
	}
	defer rows.Close()
	for rows.Next() {
		count++
	}
	return
}

//通过标签集合获取到用户关注的
func CountAccountTagsWithArray(accountId int, tags []string) (count int, err error) {
	rows, err := gormDB.Model(modelTagDomain).Select("domain_id").
		Joins("JOIN  account_domains on account_domains.id=domain_account_id ").
		Where("account_id=? AND tag_domains.tag_name in (?)", accountId, tags).Having("count(*)=?", len(tags)).Group("domain_id").Rows()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, nil
		}
		return 0, err
	}
	defer rows.Close()

	for rows.Next() {
		count++
	}

	return count, nil
}

//获取用户搜索的标签的分页数据
func SearchAccountTags(email, tag string, start, limit int) (ids []int, err error) {
	ids = make([]int, 0)
	rows, err := gormDB.Model(modelTagDomain).Select("DISTINCT domain_id").
		Joins("JOIN account_domains ON account_domains.id=domain_account_id ").
		Joins("JOIN accounts ON account_id =accounts.id ").
		Joins("JOIN domain_results ON account_domains.domain_id=domain_results.id ").
		Where("accounts.email=? AND tag_domains.tag_name like ?", email, tag).Limit(limit).Offset(start).Rows()
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return ids, nil
		}
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// 搜索添加tag的监控域名
func SearchDomainIdsOfAccountByTags(accountId int, tags []string, offset, limit int) (ids []int, total int, err error) {
	var ads []model.AccountDomain

	err = gormDB.Model(modelAccountDomain).Select("DISTINCT domain_id").
		Joins("JOIN tag_domains ON account_domains.id=domain_account_id").
		Where("account_id=? AND tag_domains.tag_name IN (?)", accountId, tags).
		Count(&total).
		Limit(limit).
		Offset(offset).
		Find(&ads).Error
	if err != nil {
		return
	}
	ids = make([]int, len(ads))
	for i, v := range ads {
		ids[i] = v.DomainId
	}
	return
}

//通过标签数组获取域名id
func SearchAccountTagsWithArray(accountId int, tags []string, offset, limit int) (ids []int, err error) {
	ids = make([]int, 0)
	rows, err := gormDB.Model(modelTagDomain).Select(" domain_id").
		Joins("JOIN account_domains ON account_domains.id=domain_account_id ").
		Joins("JOIN domain_results ON account_domains.domain_id=domain_results.id ").
		Where("account_id=? AND tag_domains.tag_name in (?)", accountId, tags).Having("count(*)=?", len(tags)).Group("domain_id").Limit(limit).Offset(offset).Rows()
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

func DelDomainAccountTags(domainAccountId int, tags []string) error {
	return gormDB.Where("domain_account_id=? AND tag_name IN (?)", domainAccountId, tags).Delete(modelTagDomain).Error
}

func AddDomainAccountTags(domainAccountId int, tags []string) (err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()
	for _, v := range tags {
		err = tx.Create(&model.TagDomain{DomainAccountId: domainAccountId, TagName: v}).Error
		if err != nil {
			return err
		}
	}
	return nil
}
