package db

import "mysslee_qcloud/model"

var modelDomainResult = &model.DomainResult{}

func UpdateDomainResult(domainId int, fields map[string]interface{}) error {
	return gormDB.Model(&model.DomainResult{}).
		Where("id=?", domainId).
		Updates(fields).Error
}
