package brand

import (
	"crypto/x509/pkix"
)

var BlackBrands []pkix.Name

func InitBlack() {
	BlackBrands = []pkix.Name{
		// WoSign
		pkix.Name{
			CommonName:   "CA 沃通根证书",
			Organization: []string{"WoSign CA Limited"},
			Country:      []string{"CN"}},
		pkix.Name{
			CommonName:   "Certification Authority of WoSign",
			Organization: []string{"WoSign CA Limited"},
			Country:      []string{"CN"},
		},
		pkix.Name{
			CommonName:   "Certification Authority of WoSign G2",
			Organization: []string{"WoSign CA Limited"},
			Country:      []string{"CN"},
		},
		pkix.Name{
			CommonName:   "CA WoSign ECC Root",
			Organization: []string{"CA WoSign ECC Root"},
			Country:      []string{"CN"},
		},

		// StartCom
		pkix.Name{
			CommonName:         "StartCom Certification Authority",
			OrganizationalUnit: []string{"Secure Digital Certificate Signing"},
			Organization:       []string{"StartCom Ltd."},
			Country:            []string{"IL"},
		},
		pkix.Name{
			CommonName:   "StartCom Certification Authority G2",
			Organization: []string{"StartCom Ltd."},
			Country:      []string{"IL"},
		},
	}
}

func IsInBlackList(pin string) bool {
	for _, b := range BlackListPins {
		if pin == b {
			return true
		}
	}
	return false
}

//判断是否在黑名单中
func IsInBlackBrand(issuer pkix.Name) bool {
	for _, black := range BlackBrands {
		//需要CN,O,OU,C四个字段都相等，
		if issuer.CommonName != black.CommonName || !checkSame(issuer.Organization, black.Organization) || !checkSame(issuer.OrganizationalUnit, black.OrganizationalUnit) || !checkSame(issuer.Country, black.Country) {
			continue
		} else {
			return true
		}
	}
	return false
}

func checkSame(issuer, black []string) bool {
	if len(issuer) != len(black) {
		return false
	}

	if (issuer == nil) != (black == nil) {
		return false
	}

	for i, j := range issuer {
		if j != black[i] {
			return false
		}
	}
	return true
}
