package model

import (
	"bytes"
	"time"

	"mysslee_qcloud/utils/certutils"
)

type Brand struct {
	Hash         string    //证书hash
	BrandGroup   string    //品牌组
	BrandName    string    //品牌名
	CertName     string    //证书Common Name
	CertBytes    []byte    //证书数据
	ConfirmedBit byte      //sqlite中的confirmed
	Confirmed    bool      //确认
	CreateTime   time.Time //创建时间
	Disabled     bool      //禁用
	Comment      string    //评论
	Kind         string
}

func (b *Brand) ConfirmedByteToBool() bool {
	if b.ConfirmedBit == 1 {
		return true
	}
	return false
}

func (b *Brand) ConfirmedToByte() byte {
	if b.Confirmed {
		return 1
	}
	return 0
}

func (b *Brand) UpdateConfirmed() {
	if b.Disabled {
		return
	}

	certs, ok := certutils.GenCertsFromPem(b.CertBytes)
	if !ok || len(certs) == 0 {
		return
	}

	cert := certs[0]
	if certutils.IsExpired(cert) {
		b.Disabled = true
		b.Comment = "过期"
	}

}

//判断两个品牌是否相等
func BrandContentEqual(one *Brand, another *Brand) bool {
	//hash不同
	if one.Hash != another.Hash {
		return false
	}

	if one.BrandGroup != another.BrandGroup {
		return false
	}

	if one.BrandGroup != another.BrandGroup {
		return false
	}

	if one.BrandName != another.BrandName {
		return false
	}

	if one.CertName != another.CertName {
		return false
	}

	if !bytes.Equal(one.CertBytes, another.CertBytes) {
		return false
	}

	if one.Confirmed != another.Confirmed {
		return false
	}

	if one.Disabled != another.Disabled {
		return false
	}

	return true
}
