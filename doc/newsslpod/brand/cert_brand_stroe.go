package brand

import (
	"crypto/x509"

	"mysslee_qcloud/utils"

	log "github.com/sirupsen/logrus"
)

var certBrandStore *CertBrandStore

func GetCertStore() *CertBrandStore {
	if certBrandStore == nil {
		log.Panic("没有初始化证书仓库")
	}

	return certBrandStore
}

// 证书仓库
type CertBrandStore struct {
	subjectKeyIdToCABrands map[string][]*CABrand //按照颁发者秘钥标识 建立索引
	caBrands               map[string]*CABrand   //证书品牌
	trustRoots             []*CABrand            //根证书
	trustCAs               []*CABrand            //中间证书
	relationCacheLen       int                   //证书缓存链长度
}

func (r *CertBrandStore) GetSubjectKeyIdToCABrands() map[string][]*CABrand {
	return r.subjectKeyIdToCABrands
}

func (r *CertBrandStore) GetLatestAuthBrand(authkeyId string) (*CABrand, bool) {
	brands, ok := r.subjectKeyIdToCABrands[authkeyId]
	if !ok {
		return nil, false
	}

	vaildBrands := []*CABrand{}
	for _, b := range brands {
		if b.Expired ||
			b.Revoke ||
			b.X509.SignatureAlgorithm == x509.SHA1WithRSA ||
			b.X509.SignatureAlgorithm == x509.DSAWithSHA1 ||
			b.X509.SignatureAlgorithm == x509.ECDSAWithSHA1 {
			continue
		}

		vaildBrands = append(vaildBrands, b)
	}

	if len(vaildBrands) == 0 {
		return nil, false
	}

	latestBrand := vaildBrands[0]
	for i := 1; i < len(vaildBrands); i++ {
		if vaildBrands[i].X509.NotBefore.Sub(latestBrand.X509.NotBefore).Seconds() > 0 {
			latestBrand = vaildBrands[i]
		}
	}
	return latestBrand, true
}

func (r *CertBrandStore) AddBrands(brands []*CABrand) {
	for _, brand := range brands {
		r.caBrands[brand.Hash] = brand
	}
}

//复制根和中间证书
func (r *CertBrandStore) GetRootsAndCAs() (trustRoots, trustCAs []*CABrand) {
	//trustRoots = make([]*CABrand, len(r.trustRoots))
	//trustCAs = make([]*CABrand, len(r.trustCAs))
	//copy(trustRoots, r.trustRoots)
	//copy(trustCAs, r.trustCAs)
	return r.trustRoots, r.trustCAs
}

//获取证书库中的根证书
func (r *CertBrandStore) GetTrustRoots() []*CABrand {
	//var trustRoots []*CABrand
	//trustRoots = make([]*CABrand, len(r.trustRoots))
	//copy(trustRoots, r.trustRoots)
	return r.trustRoots
}

///获取证书库中的中间证书
func (r *CertBrandStore) GetTrustCAs() []*CABrand {
	//var trustCAs []*CABrand
	//trustCAs = make([]*CABrand, len(r.trustCAs))
	//copy(trustCAs, r.trustCAs)
	return r.trustCAs
}

func (r *CertBrandStore) CAInTrustCAs(trustCA *x509.Certificate) bool {
	// TODO: 可考虑二分查找
	for _, cert := range r.trustCAs {
		if cert.Hash == utils.SHA1String(trustCA.Raw) {
			return true
		}
	}
	return false
}

//增加中间证书
func (r *CertBrandStore) AddTrustCA(trustCA *CABrand) {
	var find bool
	for _, i := range r.trustCAs {
		if trustCA.Hash == i.Hash {
			find = true
			break
		}
	}

	if !find {
		r.trustCAs = append(r.trustCAs, trustCA)
		r.caBrands[trustCA.Hash] = trustCA
	}
}

//获取所有的品牌证书
func (r *CertBrandStore) GetCaBrands() map[string]*CABrand {
	brands := make(map[string]*CABrand)
	for k, v := range r.caBrands {
		brands[k] = v
	}
	return brands
}
