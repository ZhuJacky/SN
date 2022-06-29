package brand

import (
	"crypto/x509"
	"crypto/x509/pkix"
	"strings"
	"time"
)

type CertInfo struct {
	SHA1      string
	Pin       string
	Issuer    pkix.Name
	NotBefore time.Time
	NotAfter  time.Time
	PubAlgo   x509.PublicKeyAlgorithm
}

type PlatformTrust struct {
	Platform string `json:"paltform"`
	Pass     bool   `json:"pass"`
	Comments string `json:"comments"`
}

//  根证书在各平台的兼容性测试
func PlatformTrustTest(chain []*CertInfo, isTrust bool) []*PlatformTrust {

	if len(chain) < 1 {
		return nil
	}

	var root, midd, cert *CertInfo
	root = chain[len(chain)-1]

	if len(chain) > 1 {
		midd = chain[len(chain)-2]
	}

	cert = chain[0]

	res := []*PlatformTrust{}
	for _, platform := range Platforms {
		pft := new(PlatformTrust)
		pft.Platform = platform
		// 针对Windows_xp Android_2_3 ECC 不支持处理
		if cert.PubAlgo == x509.ECDSA && (platform == Windows_xp || platform == Android_2_3) {
			pft.Pass = false
			res = append(res, pft)
			continue
		}

		if !isTrust {
			pft.Pass = false
		} else {
			switch platform {
			case Windows_xp, Windows_7, Windows_8, Windows_10:
				pft.Pass = IsInWinTrustStore(root, midd)

			case Firefox_51, Firefox_54:
				comm, exist := isInFirefoxBlacklist(chain)
				pft.Comments = comm
				if exist {
					pft.Pass = false
				} else {
					pft.Pass = trust(platform, root, midd)
				}

			default:
				pft.Pass = trust(platform, root, midd)
			}
		}

		res = append(res, pft)
	}
	return res
}

func trust(platform string, root, midd *CertInfo) bool {

	for _, hp := range PlatformTrustCALists[platform] {

		if strings.ToLower(root.SHA1) == strings.ToLower(hp.Hash) {
			return true
		}
		if midd != nil {
			if midd.Pin == hp.Pin {
				return true // 解决交叉问题
			}
		}
	}

	return false
}

func isInFirefoxBlacklist(chain []*CertInfo) (comment string, exist bool) {

	l := len(chain)
	rootSHA1 := chain[l-1].SHA1

	switch strings.ToUpper(rootSHA1) {
	case
		"3E2BF7F2031B96F38CE6C4D8A85D3E2D58476A0F",
		"A3F1333FE242BFCFC5D14E8F394298406810D1A0",
		"31F1FD68226320EEC63B3F9DEA4A3E537C7C3917":
		comment = "Certs issued after October 21, 2016 that chain up to StartCom CA root certificates are not trusted in Mozilla products, beginning with Firefox 51."
	case
		"D27AD2BEED94C0A13CC72521EA5D71BE8119F32B",
		"1632478D89F9213A92008563F5A4A7D312408AD6",
		"B94294BF91EA8FB64BE61097C7FB001359B676CB",
		"FBEDDC9065B7272037BC550C9C56DEBBF27894E1":
		comment = "Certs issued after October 21, 2016 that chain up to WoSign CA root certificates are not trusted in Mozilla products, beginning with Firefox 51."
	default:
		exist = false
		return
	}

	if l < 2 {
		exist = false
		return
	}

	middCert := chain[l-2]
	vaildDate, _ := time.Parse("2006-01-02", "2016-10-21")
	exist = middCert.NotBefore.After(vaildDate)
	return
}
