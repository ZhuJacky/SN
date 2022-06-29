package brand

import (
	"crypto/x509"
	"database/sql"
	"fmt"

	"mysslee_qcloud/utils/certutils"

	log "github.com/sirupsen/logrus"
)

//生成并保存证书缓存
func GenerateAndSaveCertBrandsToDb() error {
	relation, err := GenerateCertBrandsCache()
	if err != nil {
		return err
	}

	err = SaveRelationToDB(relation, false, false)
	if err != nil {
		return err
	}
	return nil
}

// 构建证书缓存
func GenerateCertBrandsCache() (relation string, err error) {
	brands, err := getCABrands()
	if err != nil {
		return "", err
	}
	roots, intermediates, err := divRootsAndIntermediates(brands)
	if err != nil {
		log.Panicf("对证书分类错误：%v", err)
	}

	checkIntermediatesCheckChain(roots, intermediates)
	return GenerateRelation(roots, intermediates)
}

//从sqlite中加载所有的所有的证书，同步ca_brand表中的差异
func getCABrands() (brands map[string]*CABrand, err error) {
	// db, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	return nil, err
	// }
	// defer db.Close()

	rows, err := dbConn.Query("SELECT hash,brand_group,brand_name,cert_name,cert_bytes FROM ca_brand WHERE  disabled <> TRUE")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	brands = make(map[string]*CABrand, 0)

	for rows.Next() {
		b := &CABrand{}
		err = rows.Scan(&b.Hash, &b.BrandGroup, &b.BrandName, &b.CA, &b.CertBytes)
		if err != nil {
			continue
		}
		b.X509, err = x509.ParseCertificate(b.CertBytes)
		if err != nil {
			log.Warnf("load brand parse cert(%v):%v", b.Hash, err)
			continue
		}
		b.CertName = certutils.CertOUString(&b.X509.Subject)
		b.Pin = certutils.GenPin(b.X509.RawSubjectPublicKeyInfo)
		b.AuthorityKeyId = fmt.Sprintf("%x", b.X509.AuthorityKeyId)
		b.SubjectKeyId = fmt.Sprintf("%x", b.X509.SubjectKeyId)

		brands[b.Hash] = b
	}

	return brands, nil
}

//更新config表数据
func updateConfig() {
	// pgdb, err := sql.Open(dbConf.Driver, dbConf.Source)
	// if err != nil {
	// 	log.Panicf("打开pg数据库错误:%v", err)
	// }
	// defer pgdb.Close()

	var id string
	err := dbConn.QueryRow("SELECT id FROM config").Scan(&id)
	switch {
	case err == sql.ErrNoRows: //没有数据插入
		_, err := dbConn.Exec("INSERT INTO  config (need_rebuild,version) VALUES (true,1)")
		if err != nil {
			log.Panicf("插入配置错误:%v", err)
		}
	case err != nil:
		log.Panicf("查找config表错误:%v", err)
	default: //有数据更新
		_, err := dbConn.Exec("UPDATE config SET need_rebuild= true, create_time=now(),version=version+1 WHERE id =$1", id)
		if err != nil {
			log.Panicf("更新配置错误:%v", err)
		}
	}
}

func insertCerCacheIntoConfig() {
	getPgCABrands()

}
