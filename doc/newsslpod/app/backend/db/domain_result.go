package db

import (
	"errors"
	"mysslee_qcloud/model"
	"mysslee_qcloud/redis"
	"time"

	"github.com/jinzhu/gorm"
)

var modelDomainResult = &model.DomainResult{}

// GetDomainResultById 通过id获取证书信息
func GetDomainResultById(domainId int) (result *model.DomainResult, err error) {
	result = &model.DomainResult{}
	err = gormDB.Model(&model.DomainResult{}).Where("id=?", domainId).First(result).Error
	if err != nil {
		return nil, err
	}
	return result, nil
}

// IsExistDomainResultReturnDomainId TODO
func IsExistDomainResultReturnDomainId(domain, ip, port string, serverType int) (int, bool) {
	var (
		count, id int
		db        *gorm.DB
	)
	if ip == "" {
		db = gormDB.Model(modelDomainResult).Select("id").
			Where("domain=? AND port=? AND server_type=? AND domain_flag&?=0", domain, port, serverType, model.DomainFlagBindIP)
	} else {
		db = gormDB.Model(modelDomainResult).Select("id").
			Where("domain=? AND ip=? AND port=? AND server_type=? AND domain_flag&?<>0", domain, ip, port, serverType,
				model.DomainFlagBindIP)
	}
	err := db.Count(&count).Row().Scan(&id)
	return id, err == nil && count > 0
}

// InsertDomainResultWithModel 通过模型插入数据
func InsertDomainResultWithModel(result *model.DomainResult) error {
	result.CreatedAt = time.Now().UTC()
	result.LastFastDetectionTime = model.TimeZeroAt
	result.LastFullDetectionTime = model.TimeZeroAt

	err := gormDB.Create(result).Error
	return err
}

// GetNeedCheckDomains 获取需要检测的域名
func GetNeedCheckDomains(t time.Time, limit func(int) int) (drs []model.DomainResult, err error) {
	// redis 并发锁
	count := 0
	for {
		count++
		if count > 5 {
			return nil, errors.New("unlocked key: need-check-domain-lock-new")
		}
		if redis.Lock("need-check-domain-lock-new") {
			break
		}
		time.Sleep(time.Millisecond * 50)
	}
	defer redis.Unlock("need-check-domain-lock-new")

	var total int

	db := gormDB.Model(modelDomainResult).
		Where("lose_efficacy=false")
	err = db.Count(&total).Error
	if err != nil {
		return
	}
	count = limit(total)
	// 获取数据
	err = db.Select(
		`id,domain,ip,punycode_domain,port,server_type,last_fast_detection_time,last_full_detection_time,prev_status,trust_status,grade,result_hash,domain_flag,domain_status`).
		Where("TIMESTAMPDIFF(MINUTE,last_fast_detection_time,NOW())>=5").
		Order("last_fast_detection_time asc").
		Limit(count).
		Find(&drs).Error
	if err != nil || len(drs) == 0 {
		return
	}

	var (
		ids []int
		now = time.Now().UTC()
	)
	for _, v := range drs {
		v.LastFastDetectionTime = now
		ids = append(ids, v.Id)
	}
	err = gormDB.Exec("UPDATE domain_results SET last_fast_detection_time=now() WHERE id IN (?)", ids).Error
	return
}

// UpdateLoseEfficacy 更新失效状态
func UpdateLoseEfficacy(id int, efficacy bool) error {
	return gormDB.Model(&model.DomainResult{}).Where("id = ?", id).Update("lose_efficacy", efficacy).Error
}

// UpdateFastCheckTime 更新快速检测的时间
func UpdateFastCheckTime(ids ...int) error {
	return gormDB.Model(&model.DomainResult{}).Where("id IN (?)", ids).Update("last_fast_detection_time", time.Now().
		UTC()).Error
}

// SearchBySecureGrade 根据等级获取域名id
func SearchBySecureGrade(accountId int, grade string, offset, limit int) (ids []int, err error) {
	ids = make([]int, 0)
	rows, err := gormDB.Model(&model.DomainResult{}).Select("domain_results.id").
		Joins("JOIN account_domains ON account_domains.domain_id = domain_results.id").
		Where("account_id=? AND domain_results.grade = ?", accountId, grade).Order("account_domains.created_at DESC").
		Offset(offset).Limit(limit).Rows()
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

// CountSearchBySecureGrade 获取
func CountSearchBySecureGrade(accountId int, grade string) (count int, err error) {
	err = gormDB.Model(&model.DomainResult{}).
		Joins("JOIN account_domains ON account_domains.domain_id = domain_results.id").
		Where("account_id=? AND domain_results.grade = ?", accountId, grade).Count(&count).Error
	if err != nil {
		if gorm.IsRecordNotFoundError(err) {
			return 0, nil
		}
		return 0, err
	}
	return count, nil
}

// GetInfoByAccountId 根据用户的邮箱获取关注的域名信息
func GetInfoByAccountId(accountId int) (results []*model.DomainResult, err error) {
	results = make([]*model.DomainResult, 0)
	err = gormDB.Model(&model.DomainResult{}).
		Joins("JOIN account_domains ON domain_id = domain_results.id").
		Where("account_id=?", accountId).Order("account_domains.id").Order("account_domains.created_at DESC ").Find(&results).
		Error
	return results, err
}

// GetDomainByBrand 根据品牌获取域名id
func GetDomainByBrand(accountId int, brand string, offset, limit int) (ids []int, err error) {
	rows, err := gormDB.Model(&model.DomainResult{}).Select("domain_results.id").
		Joins("JOIN account_domains ON account_domains.domain_id = domain_results.id").
		Where("account_id=? AND domain_status REGEXP ?", accountId, brand).Offset(offset).Limit(limit).
		Order("account_domains.created_at DESC ").Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}

// CountDomainByBrand 获取品牌的数量
func CountDomainByBrand(accountId int, brand string) (count int, err error) {
	err = gormDB.Model(&model.DomainResult{}).
		Joins("JOIN account_domains ON account_domains.domain_id = domain_results.id").
		Where("account_id=? AND domain_status REGEXP ?", accountId, brand).Count(&count).Error
	return
}

// ResetDomainDetectionTime 重置域名数据检测时间
func ResetDomainDetectionTime(domainId int, now time.Time) error {
	fastTime := model.TimeZeroAt
	fullTime := model.TimeZeroAt
	return gormDB.Model(&model.DomainResult{}).Where("id=?", domainId).
		Updates(map[string]interface{}{
			"last_fast_detection_time": fastTime,
			"last_full_detection_time": fullTime,
		}).Error
}

// NOTE
// 获取用户监控的信息
type resultInfo struct {
	DomainId int `gorm:"column:id"`
	Domain   string
	children []*resultInfo
}

// GetAccountDomainResult TODO
// func GetAccountDomainResult(accountId int, offset, limit int) (results []*model.DomainResult, total int, err error) {
// 	var infos []*resultInfo
// 	err = gormDB.Model(modelDomainResult).Select("domain_results.id,domain").
// 		Joins("JOIN account_domains ON domain_results.id=account_domains.domain_id").
// 		Where("account_id=?", accountId).Order("account_domains.id DESC").Scan(&infos).Error
// 	if err != nil {
// 		return nil, 0, err
// 	}
//
// 	// 处理域名
// 	var sorted []*resultInfo
// 	m := make(map[string]*resultInfo)
// 	for _, v := range infos {
// 		if _, ok := m[v.Domain]; !ok {
// 			sorted = append(sorted, v)
// 			m[v.Domain] = v
// 		} else {
// 			m[v.Domain].children = append(m[v.Domain].children, v)
// 		}
// 	}
//
// 	// 分页
// 	total = len(m)
// 	if offset >= total {
// 		return nil, 0, nil
// 	}
// 	end := offset + limit
// 	if end > total {
// 		end = total
// 	}
//
// 	// 获取 domain result
// 	for i := offset; i < end; i++ {
// 		domainResult, err := GetDomainResultAndChildren(accountId, sorted[i])
// 		if err != nil {
// 			return nil, 0, err
// 		}
// 		results = append(results, domainResult)
// 	}
// 	return
// }
func GetAccountDomainResult(accountId, offset, limit int) ([]*model.DomainResult, int, error) {
	var (
		results []*model.DomainResult
		infos   []*resultInfo
		total   int
	)
	err := gormDB.Model(modelDomainResult).Select("domain_results.id").
		Joins("JOIN account_domains ON domain_results.id=account_domains.domain_id").
		Where("account_id=?", accountId).
		Count(&total).
		Limit(limit).
		Offset(offset).
		Order("account_domains.id DESC").
		Scan(&infos).Error

	results = make([]*model.DomainResult, len(infos))
	for i, v := range infos {
		results[i], err = GetDomainResultWithOtherInfo(accountId, v.DomainId)
		if err != nil {
			return nil, 0, err
		}
	}
	return results, total, err
}

// GetDomainResultAndChildren 获取domainResult及其children
func GetDomainResultAndChildren(accountId int, info *resultInfo) (*model.DomainResult, error) {
	domainResult, err := GetDomainResultWithOtherInfo(accountId, info.DomainId)
	if err != nil {
		return nil, err
	}

	// load children
	if info.children != nil {
		for _, v := range info.children {
			result, err := GetDomainResultWithOtherInfo(accountId, v.DomainId)
			if err != nil {
				return nil, err
			}
			domainResult.Children = append(domainResult.Children, result)
		}
	}
	return domainResult, nil
}

// GetDomainResultWithOtherInfo 获取 domain result 包含其他信息
func GetDomainResultWithOtherInfo(accountId, domainId int) (*model.DomainResult, error) {
	domainResult, err := GetDomainResultById(domainId)
	if err != nil {
		return nil, err
	}

	// account_domains
	accountDomain, err := GetDomainAccountInfo(accountId, domainId)
	if err != nil {
		return nil, err
	}

	// tag_domains
	tags, err := GetTagsNameByDomainAccountId(accountDomain.Id)
	if err != nil {
		return nil, err
	}

	domainResult.AccountDomainId = accountDomain.Id
	domainResult.Notice = accountDomain.Notice
	domainResult.Tags = tags
	return domainResult, nil
}

// 获取 id、status
type statusInfo struct {
	DomainId int    `gorm:"column:id"`
	Status   string `gorm:"column:domain_status"`
}

// GetDomainResultStatusAndId TODO
func GetDomainResultStatusAndId(accountId int) ([]*statusInfo, error) {
	var result []*statusInfo
	err := gormDB.Model(modelDomainResult).Select("domain_results.id,domain_status").
		Joins("JOIN account_domains ON domain_id=domain_results.id").
		Where("account_id=?", accountId).
		Order("account_domains.id DESC").
		Scan(&result).Error
	return result, err
}

// GetByDomain 根据域名获取域名id
func GetByDomain(accountId int, punycodeDomain string) (ids []int, err error) {
	rows, err := gormDB.Model(&model.DomainResult{}).Select("domain_results.id").
		Joins("JOIN account_domains ON account_domains.domain_id = domain_results.id").
		Where("account_id=? AND punycode_domain like ?", accountId, punycodeDomain+"%").Rows()
	if err != nil {
		return nil, err
	}

	defer rows.Close()
	for rows.Next() {
		var id int
		err := rows.Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}
	return ids, nil
}
