package core

import (
	"context"
)

type CertNode struct {
	Hash         string    //证书的hash
	CertFromSSL2 bool      //证书从ssl2获取
	Cert         *OutCerts //证书内容
}

//判断证书是否已经存在
func CertNodeIsExist(nodes []*CertNode, hash string) bool {
	for _, v := range nodes {
		if hash == v.Hash {
			return true
		}
	}
	return false
}

//添加证书
func AddCertNode(ctx context.Context, nodes []*CertNode, cert *OutCerts) []*CertNode {
	node := &CertNode{
		Hash: cert.CertsInfo[0].Sha1,
		Cert: cert,
	}

	if !cert.DoCheckOCSP { //没有做ocsp检测
		cert.CheckOCSP(ctx)
	}

	nodes = append(nodes, node)
	return nodes
}
