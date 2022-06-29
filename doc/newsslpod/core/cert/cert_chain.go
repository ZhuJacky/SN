package cert

import (
	"context"
	"crypto/sha1"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"mysslee_qcloud/brand"
	"mysslee_qcloud/utils/certutils"
)

type Chain struct {
	RSA struct {
		ID    string `json:"id"`
		Chain string `json:"chain"`
	} `json:"rsa"`
	ECC struct {
		ID    string `json:"id"`
		Chain string `json:"chain"`
	} `json:"ecc"`
}

func GetOneCertChain(ctx context.Context, cert *x509.Certificate) (chain []byte, err error) {
	chains, ok := getCertChainFromCABrands(cert)
	if !ok {
		chains, err = GetCertChainFromIssuerURL(ctx, cert)
		if err != nil {
			return nil, err
		}
	}

	chain = EncodeChain(chains)
	return
}

func GetCertChain(ctx context.Context, certPEMs [][]byte) (*Chain, error) {
	chain := &Chain{}
	for _, pem := range certPEMs {
		deCert, err := certutils.DecodeCert(pem)
		if err != nil {
			return nil, err
		}

		enChain, err := GetOneCertChain(ctx, deCert)
		if err != nil {
			return nil, err
		}
		chainId := fmt.Sprintf("%X", sha1.Sum(enChain))
		enChainStr := string(enChain)

		switch deCert.PublicKeyAlgorithm {
		case x509.RSA:
			chain.RSA.ID = chainId
			chain.RSA.Chain = enChainStr
		case x509.ECDSA:
			chain.ECC.ID = chainId
			chain.ECC.Chain = enChainStr
		}
	}
	return chain, nil
}

func EncodeChain(chain []*x509.Certificate) []byte {
	var chains []byte
	for i := 0; i < len(chain); i++ {
		cert := chain[i]
		block := pem.Block{
			Type:  "CERTIFICATE",
			Bytes: cert.Raw,
		}
		certPEM := pem.EncodeToMemory(&block)
		chains = append(chains, certPEM...)
	}
	return chains
}

// 从缓存中组链
// 策略修改 按照颁发者秘钥标识到缓存中查找 排除sha1
func getCertChainFromCABrands(cert *x509.Certificate) ([]*x509.Certificate, bool) {
	certs := []*x509.Certificate{cert}

	authKeyId := fmt.Sprintf("%X", cert.AuthorityKeyId)
	for {
		// 按照颁发者ID 找到颁发者
		cabrand, ok := brand.GetCertStore().GetLatestAuthBrand(authKeyId)
		if !ok {
			return nil, false
		}

		// 是根证书
		if cabrand.IsRoot {
			return certs, true
		}

		certs = append(certs, cabrand.X509)
		authKeyId = cabrand.AuthorityKeyId
	}
}

//从证书中的caUrl中获取上级证书
func GetCertChainFromIssuerURL(ctx context.Context, cert *x509.Certificate) (certs []*x509.Certificate, err error) {

	certs = append(certs, cert)
	for certs[len(certs)-1].IssuingCertificateURL != nil {

		parsentURL := certs[len(certs)-1].IssuingCertificateURL[0]
		cert, err = certutils.GetCertFromIssuerURL(ctx, parsentURL)

		if err != nil {
			break
		}

		if isSelfSigned(cert) {
			break
		}

		certs = append(certs, cert)
	}
	return certs, nil

}

func isSelfSigned(cert *x509.Certificate) bool {
	return cert.CheckSignatureFrom(cert) == nil
}
