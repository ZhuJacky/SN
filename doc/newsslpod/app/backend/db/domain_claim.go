package db

import "mysslee_qcloud/model"

// DomainClaimList 查询域名认领列表
func DomainClaimList(accountId, page, pageSize int) ([]*model.DomainClaim, int, error) {
	var (
		total      int
		domainList []*model.DomainClaim
	)
	err := gormDB.Model(model.DomainClaim{}).
		Where("account_id = ?", accountId).
		Count(&total).
		Order("id DESC").
		Offset((page - 1) * pageSize).
		Limit(pageSize).
		Find(&domainList).Error
	return domainList, total, err
}

// IsExistDomainClaim 查看认领域名是否存在
func IsExistDomainClaim(accountId int, domain string) bool {
	var count int
	err := gormDB.Model(model.DomainClaim{}).Where("domain = ? AND account_id = ?", domain, accountId).Count(&count).Error
	return err == nil && count > 0
}

// IsDomainClaimed 查看是否已经认领过
func IsDomainClaimed(accountId int, domain string) bool {
	var count int
	err := gormDB.Model(model.DomainClaim{}).Where("domain = ? AND account_id = ? AND status = ?", domain, accountId, model.DomainClaimStatusVerified).Count(&count).Error
	return err == nil && count > 0
}

// GetDomainClaimByAccountIdAndDomain 通过用户ID及域名查询认领域名
func GetDomainClaimByAccountIdAndDomain(accountId int, domain string) ([]*model.DomainClaim, error) {
	var domainClaim []*model.DomainClaim
	err := gormDB.Model(model.DomainClaim{}).Where("account_id = ? AND domain = ?", accountId, domain).First(&domainClaim).Error
	return domainClaim, err
}

// DomainClaimAdd 添加认领域名
func DomainClaimAdd(domainClaim *model.DomainClaim) error {
	err := gormDB.Create(domainClaim).Error
	return err
}

// GetDomainClaimByIdAndAccountId 通过id和用户id获取认领域名
func GetDomainClaimByIdAndAccountId(id, accountId int) (*model.DomainClaim, error) {
	domainClaim := new(model.DomainClaim)
	err := gormDB.Model(model.DomainClaim{}).Where("id = ? AND account_id = ?", id, accountId).First(domainClaim).Error
	return domainClaim, err
}

// DomainClaimStatusUp 更改认领域名状态
func DomainClaimStatusUp(id, status int) error {
	err := gormDB.Model(model.DomainClaim{}).Where("id = ?", id).Update("status", status).Error
	return err
}

// DomainClaimDel 删除认领域名
func DomainClaimDel(id, accountId int) (int, error) {
	var count int
	err := gormDB.Model(&model.DomainClaim{}).Where("id = ? AND account_id = ?", id, accountId).Count(&count).Delete(&model.DomainClaim{}).Error
	return count, err
}
