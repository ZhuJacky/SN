package db

import (
	"mysslee_qcloud/model"
	"time"
)

// GetDomainRegionalResult 根据域名ID和地域获得
func GetDomainRegionalResult(domainID int, region string) (*model.DomainRegionalResult, error) {
	ret := &model.DomainRegionalResult{}
	err := gormDB.Model(&model.DomainRegionalResult{}).Where("domain_id = ?", domainID).
		Where("region = ?", region).Limit(1).Find(ret).Error
	return ret, err
}

// UpdateDomainRegionalResult 更新地域检测信息
func UpdateDomainRegionalResult(drr *model.DomainRegionalResult) error {
	err := gormDB.Model(&model.DomainRegionalResult{}).Where("domain_id=?", drr.DomainID).Where("region=?", drr.Region).Update(drr).Error
	return err
}

// InsertDomainRegionalResult 插入地域检测信息
func InsertDomainRegionalResult(drr *model.DomainRegionalResult) error {
	err := gormDB.Create(drr).Error
	return err
}

// UpdateDomainRegionalResultDetectionTime 更新地域检测的时间
func UpdateDomainRegionalResultDetectionTime(domainID int, region string) error {
	err := gormDB.Model(&model.DomainRegionalResult{}).Where("domain_id=?", domainID).Where("region=?", region).Update("last_detection_time", time.Now().Local()).Error
	return err
}
