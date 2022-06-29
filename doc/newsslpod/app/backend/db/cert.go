// Package db provides ...
package db

import (
	"mysslee_qcloud/model"
)

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

// 获取域名相关联的证书信息
func GetDomainCertInfo(domainId int) ([]*model.CertInfo, error) {
	var cis []*model.CertInfo
	err := gormDB.Model(modelCertInfo).
		Joins("JOIN domain_certs ON domain_certs.hash=cert_infos.hash").
		Where("domain_id=?", domainId).
		Find(&cis).Error
	return cis, err
}

// 获取证书列表
var SupportedSearch = map[string]bool{
	"domainId":   true,
	"commonName": true,
	"hash":       true,
}

func GetDomainCertList(aid, offset, limit int, search string, target interface{}) ([]*model.CertInfo, int, error) {
	var (
		total int
		infos []*model.CertInfo
	)

	infoDB := gormDB.Model(modelCertInfo).
		Joins("JOIN domain_certs ON domain_certs.hash=cert_infos.hash").
		Joins("JOIN account_domains ON account_domains.domain_id=domain_certs.domain_id").
		Where("account_id=?", aid)
	switch search {
	case "domainId":
		infoDB = infoDB.Where("account_domains.domain_id=?", target)
	case "commonName":
		infoDB = infoDB.Where("cn=?", target)
	case "hash":
		infoDB = infoDB.Where("domain_certs.hash=?", target)
	}
	err := infoDB.Select("COUNT(DISTINCT domain_certs.hash,domain_certs.trust_status)").Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = infoDB.Select("DISTINCT cert_infos.*,domain_certs.trust_status").
		Order("end_time ASC").
		Limit(limit).
		Offset(offset).
		Find(&infos).Error
	if err != nil {
		return nil, 0, err
	}
	return infos, total, nil
}

// 获取该域名的相关证书
func GetDomainCertDetail(aid, domainId int) ([]*model.CertInfo, error) {
	var infos []*model.CertInfo
	err := gormDB.Select("cert_infos.*,domain_certs.trust_status").
		Joins("JOIN domain_certs ON domain_certs.hash=cert_infos.hash").
		Joins("JOIN account_domains ON account_domains.domain_id=domain_certs.domain_id").
		Where("account_domains.account_id=? AND domain_certs.domain_id=?", aid, domainId).
		Order("cert_infos.created_at ASC").
		Find(&infos).Error
	return infos, err
}

// 获取该 hash 相关联的域名
func GetDomainByHash(aid int, hash string, offset, limit int) ([]int, int, error) {
	var total int

	infoDB := gormDB.Model(modelDomainCert).
		Joins("JOIN account_domains ON account_domains.domain_id=domain_certs.domain_id").
		Where("account_id=? AND hash=?", aid, hash)
	err := infoDB.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}
	rows, err := infoDB.Select("domain_certs.domain_id").
		Order("account_domains.created_at").
		Limit(limit).
		Offset(offset).Rows()
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		err = rows.Scan(&id)
		if err != nil {
			return nil, 0, err
		}
		ids = append(ids, id)
	}
	return ids, total, nil
}

// GetCertInfoByHash 根据hash获得证书内容
func GetCertInfoByHash(hash string) (*model.CertInfoShow, error) {
	ret := []*model.CertInfo{}
	err := gormDB.Model(&model.CertInfo{}).Where("hash=?", hash).Find(&ret).Error
	if len(ret) != 0 {
		return model.CertInfoForShow(ret)[0], err
	}
	return &model.CertInfoShow{}, err
}
