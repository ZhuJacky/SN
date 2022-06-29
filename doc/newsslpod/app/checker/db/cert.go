// Package db provides ...
package db

import "mysslee_qcloud/model"

var (
	modelCertInfo   = &model.CertInfo{}
	modelDomainCert = &model.DomainCert{}
)

func IsExistCertInfo(hash string) bool {
	var count int
	err = gormDB.Model(modelCertInfo).Where("hash=?", hash).Count(&count).Error
	return err == nil && count > 0
}

// 添加证书
func AddCertInfo(info *model.CertInfo) error {
	return gormDB.Create(info).Error
}

// 更新域名证书关系
func UpDomainCert(domainId int, delDcs, upDcs, newDcs []*model.DomainCert) (err error) {
	tx := gormDB.Begin()
	defer func() {
		if err != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	for _, v := range delDcs {
		err = tx.Delete(v).Error
		if err != nil {
			return
		}
	}

	for _, v := range upDcs {
		err = tx.Model(modelDomainCert).
			Where("domain_id=? AND hash=?", domainId, v.Hash).
			Update("trust_status", v.TrustStatus).Error
		if err != nil {
			return
		}
	}

	for _, v := range newDcs {
		err = tx.Create(v).Error
		if err != nil {
			return
		}
	}
	return
}

// 获取域名关联的证书
func GetDomainCert(domainId int) ([]*model.DomainCert, error) {
	var dcs []*model.DomainCert
	err := gormDB.Where("domain_id=?", domainId).Find(&dcs).Error
	return dcs, err
}
