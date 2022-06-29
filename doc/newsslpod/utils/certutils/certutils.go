package certutils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/base64"
	"encoding/pem"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"strings"
	"time"

	"mysslee_qcloud/core/myconn"

	"github.com/fullsailor/pkcs7"
	"golang.org/x/net/idna"
)

var ErrInvaildCert = errors.New("invaild certificate")

//验证根证书   处理investor.qualys.com 服务端证书中存在一张MD2的根证书
func CheckRoot(subject *x509.Certificate) bool {
	result, err := checkSignFrom(subject, subject)
	if err == nil {
		return result
	}
	switch err.(type) {
	case x509.InsecureAlgorithmError:
		return true
	}

	if strings.Contains(err.Error(), "cannot verify signature: algorithm unimplemented") {
		return true
	}
	return false
}

func CheckSignFrom(subject, issuer *x509.Certificate) bool {
	result, err := checkSignFrom(subject, issuer)
	if err == nil {
		return result
	}

	switch err.(type) {
	case x509.InsecureAlgorithmError:
		return true
	}
	return false
}

func checkSignFrom(subject, issuer *x509.Certificate) (result bool, err error) {
	if !checkSignFromUseCertInfo(subject, issuer) {
		return false, nil
	}
	if err = subject.CheckSignatureFrom(issuer); err == nil {
		return true, nil
	} else {
		return true, err
	}
}

//用证书基础信息进行签名验证(性能优化)
func checkSignFromUseCertInfo(subject, issuer *x509.Certificate) bool {
	//1. 验证subject的Issuer信息，和issuer的Subject信息是否相等
	if bytes.Equal(subject.RawIssuer, issuer.RawSubject) {
		return true
	}

	//因为let's encrypted 的中间证书	Let's Encrypt Authority X1 和 	Let's Encrypt Authority X3 使用相同的SubjectId 所有注释使用KeyId进行验证
	////2. 验证Sujbect的AuthorityKeyId和issuer的Subject信息是否相等
	//if bytes.Equal(subject.AuthorityKeyId, issuer.SubjectKeyId) {
	//	return true
	//}

	//3. 验证issuer的公钥使用范围（不用实现 go源码中已经有做了）
	return false
}

//DerCertToPEM 生成证书的PEM内容
func DerCertToPEM(raw []byte) string {
	var t pem.Block
	t.Bytes = raw
	t.Type = "CERTIFICATE"
	a := pem.EncodeToMemory(&t)
	return string(a)
}

//CertOUString 输出信息
func CertOUString(info *pkix.Name) string {
	if info.CommonName != "" {
		return info.CommonName
	}

	if len(info.OrganizationalUnit) != 0 {
		return info.OrganizationalUnit[0]
	}

	if len(info.Organization) != 0 {
		return info.Organization[0]
	}
	return ""

}

func GenPublicKeyFromPem(pubkey []byte) ([]byte, bool) {
	for len(pubkey) > 0 {
		var block *pem.Block
		block, pubkey = pem.Decode(pubkey)
		if block == nil {
			break
		}
		if block.Type != "PUBLIC KEY" {
			continue
		}
		return block.Bytes, true
	}
	return nil, false
}

//根据公钥生成Pin
func GenPin(publickey []byte) string {
	pin256 := sha256.Sum256(publickey)
	data := pin256[:]
	pin := base64.StdEncoding.EncodeToString(data)
	return pin
}

func UnicodeDomain(san []string) []string {
	var result = make([]string, 0)
	for i := 0; i < len(san); i++ {
		ad, err := idna.ToUnicode(san[i])
		if err != nil {
			result = append(result, san[i])
		} else {
			result = append(result, ad)
		}
	}
	return result
}

// ip to string
func IPAddressesToStr(ips []net.IP) []string {
	if ips == nil {
		return nil
	}

	strs := make([]string, len(ips))
	for i, v := range ips {
		strs[i] = v.String()
	}
	return strs
}

// 检测是否有ocsp_must_staple
func CheckOCSPMustStaple(cert *x509.Certificate) bool {
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(asn1.ObjectIdentifier{1, 3, 6, 1, 5, 5, 7, 1, 24}) {
			return true
		}
	}
	return false
}

//从证书的IssuerUrl中获取该证书的签发证书
func GetCertFromIssuerURL(ctx context.Context, issuerUrl string) (cert *x509.Certificate, err error) {
	req, err := http.NewRequest(http.MethodGet, issuerUrl, nil)
	if err != nil {
		return nil, err
	}
	req = req.WithContext(ctx)

	client := &http.Client{
		Transport: myconn.TransportSkipVerify,
	}

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("down cert statuscode=%v", resp.StatusCode)
	}

	data, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return DecodeCert(data)
}

//生成证书
func DecodeCert(data []byte) (*x509.Certificate, error) {
	if IsPEM(data) {
		block, _ := pem.Decode(data)
		if block == nil {
			return nil, ErrInvaildCert
		}
		if block.Type != "CERTIFICATE" {
			return nil, ErrInvaildCert
		}
		data = block.Bytes
	}

	cert, err := x509.ParseCertificate(data)
	if err == nil {
		return cert, nil
	}

	p, err := pkcs7.Parse(data)
	if err == nil {
		return p.Certificates[0], nil
	}

	return nil, ErrInvaildCert
}

func IsPEM(data []byte) bool {
	return bytes.HasPrefix(data, []byte("-----BEGIN "))
}

func IsExpired(cert *x509.Certificate) bool {
	return cert.NotBefore.UTC().Sub(time.Now().UTC()) > 0 || cert.NotAfter.UTC().Sub(time.Now().UTC()) < 0
}

//GenCertsFromPem 从pem内容中获取多个证书
func GenCertsFromPem(pemCerts []byte) (certs []*x509.Certificate, ok bool) {
	for len(pemCerts) > 0 {
		var block *pem.Block
		block, pemCerts = pem.Decode(pemCerts)
		if block == nil {
			break
		}
		if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
			continue
		}

		cert, err := x509.ParseCertificate(block.Bytes)
		if err != nil {
			continue
		}

		certs = append(certs, cert)
		ok = true
	}
	return
}

//GenCertFromPem 从pem中获取单张证书
func GenCertFromPem(pemCert []byte) (leaf *x509.Certificate, ok bool) {
	var block *pem.Block
	block, pemCert = pem.Decode(pemCert)
	if block == nil {
		return
	}
	if block.Type != "CERTIFICATE" || len(block.Headers) != 0 {
		return
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return
	}
	return cert, true
}
