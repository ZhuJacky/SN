//Package core 处理证书相关信息
package cert

import (
	"context"
	"crypto/dsa"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"errors"
	"fmt"
	"strings"
	"time"

	"mysslee_qcloud/brand"
	"mysslee_qcloud/core/ct"
	"mysslee_qcloud/utils"
	"mysslee_qcloud/utils/certutils"

	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
)

type TrustStatus byte

const (
	Untrusted      TrustStatus = iota //不可信
	Trust                             //可信
	TrustButMissCA                    //可信但是缺链
	BlackList                         //在黑名单中
)

const (
	DV      = 1
	OV      = 2
	EV      = 3
	NoAudit = 4
)

func CertTypeToString(typ int) string {
	switch typ {
	case 1:
		return "DV"
	case 2:
		return "OV"
	case 3:
		return "EV"
	case 4:
		return "NoAudit"
	}

	return "unknown"
}

const (
	ChainRootFromServer byte = 0x04
	ChainAdditionCA     byte = 0x02
)

//OCSPInfo OCSP信息
type OCSPInfo struct {
	Status           int       `json:"status"`
	ProducedAt       time.Time `json:"produced_at"`
	NextUpdate       time.Time `json:"next_update"`
	RevokedAt        time.Time `json:"revoked_at"`
	RevocationReason int       `json:"revocation_reason"`
}

//CertInfo 证书信息
type CertInfo struct {
	SignatureAlgorithm string    //签名算法
	KeyType            string    //公钥类型
	KeyLong            int       //公钥长度
	NotAfter           time.Time //有效期
	NotBefore          time.Time //有效期
	SN                 string    //序列号
	Sha1               string    //sha1指纹
	Subject            string    //项目
	Organization       string    //组织
	OrganizationUnit   string    //组织部门
	Issuer             string    //签发者
	CertPem            string    //pem内容
	Pin                string    //公钥Pin
	AIA                []string  //AIA
	OCSPServer         []string  //Online Certificate Status Protocol
	IsRoot             bool      //是否是根证书

	KeyTypeRaw x509.PublicKeyAlgorithm
	SignAlgo   x509.SignatureAlgorithm
	X509       *x509.Certificate //证书内容

	Order    int  //服务器端的顺序
	IsServer bool //是否是服务器端发送的

	// SM2 证书特有的
	IsSM2         bool
	SM2SignAlgo   sm2.SignatureAlgorithm
	SM2KeyTypeRaw sm2.PublicKeyAlgorithm
	SM2           *sm2.Certificate
}

func (c *CertInfo) String() string {
	return fmt.Sprintf("SignatureAlgorithm :%v,KeyType:%v,NotAfter:%v,NotBefore:%v,SN:%v,sha1:%v,Subject:%v,Organization:%v,"+
		"OrganizationUnit:%v,Issuer:%v,pin:%v,order:%v,IsServer:%v,isRoot:%v", c.SignatureAlgorithm, c.KeyType, c.NotAfter, c.NotBefore, c.SN, c.Sha1, c.Subject, c.Organization, c.OrganizationUnit, c.Issuer, c.Pin, c.Order, c.IsServer, c.IsRoot)
}

type ChainResult struct {
	Certs       []*CertInfo
	CertStatus  TrustStatus
	ChainStatus byte
}

func (t TrustStatus) IsTrust() bool {
	if t != Untrusted && t != BlackList { //如果不可信，或者在黑名单中
		return true
	}

	return false
}

//CertsRawParseCerts 把服务器的原始证书链信息转换成证书
func CertsRawParseCerts(data []byte) (certs []*x509.Certificate, certserr error) {
	var i int
	datal := len(data)
	for i < datal {
		n := int(data[i])
		n <<= 8
		i++
		n += int(data[i])
		n <<= 8
		i++
		n += int(data[i])
		i++
		if len(data) < i+n {
			return nil, errors.New("获取证书长度错误")
		}
		cert, err := x509.ParseCertificate(data[i : i+n])
		if err != nil {
			certserr = err
			return
		}
		certs = append(certs, cert)
		i += n
	}
	return
}

//CertToCertInfo 把证书生成自定义的证书信息
func CertToCertInfo(certs []*x509.Certificate, isServer bool, order []*x509.Certificate) []*CertInfo {
	var certInfos []*CertInfo
	for _, cert := range certs {
		info := &CertInfo{
			SN:                 fmt.Sprintf("%X", cert.SerialNumber),
			Sha1:               utils.SHA1String(cert.Raw),
			NotAfter:           cert.NotAfter,
			NotBefore:          cert.NotBefore,
			SignAlgo:           cert.SignatureAlgorithm,
			SignatureAlgorithm: GetSignatureAlgorithm(cert.SignatureAlgorithm),
			KeyTypeRaw:         cert.PublicKeyAlgorithm,
			KeyType:            GetKeyType(cert.PublicKeyAlgorithm),
			Subject:            certutils.CertOUString(&cert.Subject),
			Issuer:             certutils.CertOUString(&cert.Issuer),
			KeyLong:            GetKeyLong(cert.PublicKey),
			CertPem:            certutils.DerCertToPEM(cert.Raw),
			Pin:                certutils.GenPin(cert.RawSubjectPublicKeyInfo),
			AIA:                cert.IssuingCertificateURL, //证书签发者url
			OCSPServer:         cert.OCSPServer,
			X509:               cert,
			IsServer:           isServer,
		}
		info.Organization, info.OrganizationUnit = certOrganization(cert.Subject)
		certInfos = append(certInfos, info)

		root := certutils.CheckRoot(cert)
		info.IsRoot = root
		// certInfos[i].AIA = cert.IssuingCertificateURL[0]
		if order != nil {
			for i := 0; i < len(order); i++ {
				if cert == order[i] {
					info.Order = i + 1
					break
				}
			}
		}
	}
	return certInfos
}

func checkTrustByChain(ctx context.Context, certs []*x509.Certificate) ([]*x509.Certificate, TrustStatus) {
	var status = Trust
	//先验证传入的证书链是否可信
	for _, c := range certs {
		if brand.IsInBlackList(certutils.GenPin(c.RawSubjectPublicKeyInfo)) { //判断是否是在黑名单中
			status = BlackList
		}
	}

	extra, t := checkTrusted(ctx, certs)
	if status == BlackList {
		return extra, BlackList
	} else {
		return extra, t
	}
}

//验证是否可信 （0：不可信，1：可信， 2：补链可信）（从池中验签名）
func checkTrusted(ctx context.Context, certs []*x509.Certificate) (additionCerts []*x509.Certificate, status TrustStatus) {
	last := certs[len(certs)-1]
	trustRoots, trustCAs := brand.GetCertStore().GetRootsAndCAs()

	for _, root := range trustRoots {
		//判断证书在组链验证中是否可以使用， 吊销，过期的证书都不再参与到组链过程
		if !brand.BrandCanUseInVerify(root) {
			continue
		}

		if certutils.CheckSignFrom(last, root.X509) {
			if brand.IsInBlackList(certutils.GenPin(root.X509.RawSubjectPublicKeyInfo)) {
				status = BlackList
				return []*x509.Certificate{root.X509}, BlackList
			}
			status = Trust
			return []*x509.Certificate{root.X509}, Trust
		}
	}

	//验证中间证书
	for _, intermediate := range trustCAs {

		//判断证书在组链验证中是否可以使用
		if !brand.BrandCanUseInVerify(intermediate) {
			continue
		}

		if certutils.CheckSignFrom(last, intermediate.X509) {
			certs := make([]*x509.Certificate, 0)
			certs = append(certs, intermediate.X509)

			finish := false

			for !finish {
				signFrom := intermediate.SignFrom
				if len(signFrom) == 0 {
					finish = true
				} else {
					sign := signFrom[0]
					if sign.IsRoot == true {
						finish = true
						certs = append(certs, sign.X509) //如果已经是根证书了，结束
						break
					} else {
						intermediate = sign //中间证书就变成刚才签发的证书
					}
				}
			}
			for _, c := range certs { //遍历判断是否在黑名单中
				if brand.IsInBlackList(certutils.GenPin(c.RawSubjectPublicKeyInfo)) {
					status = BlackList
					return certs, BlackList
				}
			}
			status = TrustButMissCA
			return certs, TrustButMissCA
		}
	}
	//收集未收集的证书
	//brand.AutoCollectUncollectedCertificate(cert)
	status = Untrusted
	return nil, Untrusted
}

//得到完整的可信链 ，存在交叉证书的问题，取得尽可能多的链
func GetFullChain(ctx context.Context, certs []*x509.Certificate) (serverCerts []*x509.Certificate, chainResult *ChainResult, err error) {
	chainResult = &ChainResult{}

	var serveradditional = make([]*x509.Certificate, 0)
	var checkedServerRoot *x509.Certificate
	var additionalRoots = make([]*x509.Certificate, 0)

	if len(certs) == 0 {
		return nil, nil, errors.New("在获取完整的证书链时，没有获取到证书")
	}

	//复制一份 用于后面的顺序判断
	serverOrders := make([]*x509.Certificate, len(certs))
	copy(serverOrders, certs)

	checkCert := certs[0]

	out := make([]*x509.Certificate, 0)
	out = append(out, checkCert)
	temp := certs[1:] //移除叶子证书

	roots, intermediates := SeparateRootAndImIntermediate(temp)
	//自签名
	signSelf := certutils.CheckRoot(checkCert)
	if signSelf {
		chainResult.Certs = CertToCertInfo(out, true, serverOrders)
		chainResult.CertStatus = Untrusted
		return certs, chainResult, nil
	}

	//进行中间证书验签
	var finish bool
	for !finish {
		var index int
		var have bool
		var cert *x509.Certificate
		for i, c := range intermediates {
			if certutils.CheckSignFrom(checkCert, c) {
				have = true
				if c.NotBefore.UTC().Sub(time.Now().UTC()) > 0 || c.NotAfter.UTC().Sub(time.Now().UTC()) < 0 {
					have = false
				} else {
					index = i
					cert = c
				}
			}

		}
		if have {
			out = append(out, cert)
			checkCert = cert
			intermediates = append(intermediates[:index], intermediates[index+1:]...)

			if len(out) > 9 {
				return nil, nil, errors.New("证书链过长")
			}

		} else {
			finish = true
		}
		if len(intermediates) == 0 {
			finish = true
		}
	}

	var hasRoot bool
	var checkRoots = make([]*x509.Certificate, 0)

	//验根
	for i, c := range roots {
		if certutils.CheckSignFrom(out[len(out)-1], c) {
			hasRoot = true
			checkedServerRoot = c
		} else {
			checkRoots = append(checkRoots, roots[i]) //其他中间证书
		}
	}

	// 额外的证书
	serveradditional = append(intermediates, checkRoots...)

	if len(serveradditional) > 0 {
		chainResult.ChainStatus = chainResult.ChainStatus | ChainAdditionCA
	}

	for _, additional := range serveradditional {
		if certutils.CheckRoot(additional) {
			additionalRoots = append(additionalRoots, additional)
		}
	}

	//可信库验证
	var trustChain []*x509.Certificate
	trustChain, chainResult.CertStatus = checkTrustByChain(ctx, out)

	if len(trustChain) == 0 { //没有额外的证书
		if hasRoot {
			out = append(out, intermediates...) //把根证书重新添加
			out = append(out, roots...)         //把根证书重新添加
		}
		chainResult.Certs = CertToCertInfo(out, true, serverOrders)
		return certs, chainResult, nil
	}

	serverChainInfo := CertToCertInfo(out, true, serverOrders) //serverChain都是服务器证书

	trustChainInfo := CertToCertInfo(trustChain[:], false, nil) //不需要剔除第一张

	if hasRoot { //没有确中间链
		chainResult.ChainStatus |= ChainRootFromServer
		//只有一个根证书的情况下
		if len(trustChain) == 1 {
			trustChainInfo = CertToCertInfo([]*x509.Certificate{checkedServerRoot}, true, serverOrders)

		} else {
			rootInfo := CertToCertInfo([]*x509.Certificate{checkedServerRoot}, true, serverOrders)

			trustChainInfo[len(trustChainInfo)-1] = rootInfo[0]
		}
	} else if len(additionalRoots) > 0 { //有额外的根证书
		if len(trustChain) > 1 {
			for _, root := range additionalRoots {
				if certutils.CheckSignFrom(trustChain[len(trustChain)-1], root) { //处理缺少中间证书，但是部署了根证书的情况
					if len(additionalRoots) == 1 && len(additionalRoots) == len(serveradditional) {
						chainResult.ChainStatus = chainResult.ChainStatus & ^ChainAdditionCA
					}
					chainResult.ChainStatus = chainResult.ChainStatus | ChainRootFromServer
					rootInfo := CertToCertInfo([]*x509.Certificate{root}, true, serverOrders)

					trustChainInfo[len(trustChainInfo)-1] = rootInfo[0]
					break
				}
			}
		}
	}
	chainResult.Certs = append(serverChainInfo, trustChainInfo...)
	return certs, chainResult, nil
}

// 查询证书透明计划
func IsTransparency(ctx context.Context, certs []*x509.Certificate, sctFromTLS []byte) (isQualified int, desc string) {
	if len(certs) < 1 {
		return
	}

	certRaws := [][]byte{}
	for _, cert := range certs {
		certRaws = append(certRaws, cert.Raw)
	}

	ctInfo, err := ct.IsCTQualified(sctFromTLS, certRaws)
	if err != nil {
		log.WithFields(log.Fields{
			"req": utils.GetReqInfoFromContext(ctx),
		}).Warnf("sct err:%v", err.Error())
		return
	}

	return ctInfo.IsQualified, ctInfo.Description
}

//证书组织
func certOrganization(info pkix.Name) (organization, unit string) {
	if len(info.Organization) > 0 {
		organization = info.Organization[0]
	}
	if len(info.OrganizationalUnit) > 0 {
		unit = info.OrganizationalUnit[0]
	}
	return
}

//获取公钥算法
func GetKeyType(publicAlgo x509.PublicKeyAlgorithm) string {
	var keyAlgo string
	switch publicAlgo {
	case x509.RSA:
		keyAlgo = "RSA"
	case x509.DSA:
		keyAlgo = "DSA"
	case x509.ECDSA:
		keyAlgo = "ECDSA"
	default:
		keyAlgo = "Unknown"
	}
	return keyAlgo
}

//GetKeyLong 获取公钥的长度
func GetKeyLong(publicKey interface{}) int {
	var len int
	switch pub := publicKey.(type) {
	case *rsa.PublicKey:
		len = pub.N.BitLen()
	case *dsa.PublicKey:
		len = pub.Y.BitLen()
	case *ecdsa.PublicKey:
		len = pub.Curve.Params().BitSize
	}
	return len
}

func GetSimpleSignatureAlgorithm(signAlgo x509.SignatureAlgorithm) string {
	var simple string
	switch signAlgo {
	case x509.MD2WithRSA:
		simple = "MD2"
	case x509.MD5WithRSA:
		simple = "MD5"
	case x509.SHA1WithRSA:
		simple = "SHA1"
	case x509.SHA256WithRSA:
		simple = "SHA256"
	case x509.SHA384WithRSA:
		simple = "SHA384"
	case x509.SHA512WithRSA:
		simple = "SHA512"
	case x509.DSAWithSHA1:
		simple = "SHA1"
	case x509.DSAWithSHA256:
		simple = "SHA256"
	case x509.ECDSAWithSHA1:
		simple = "SHA1"
	case x509.ECDSAWithSHA256:
		simple = "SHA256"
	case x509.ECDSAWithSHA384:
		simple = "SHA384"
	case x509.ECDSAWithSHA512:
		simple = "SHA512"
	default:
		simple = "Unknown"
	}
	return simple
}

//获取签名算法
func GetSignatureAlgorithm(signAlgo x509.SignatureAlgorithm) string {
	var signAlgoStr string
	switch signAlgo {
	case x509.UnknownSignatureAlgorithm:
		signAlgoStr = "Unknown"
	case x509.MD2WithRSA:
		signAlgoStr = "MD2WithRSA"
	case x509.MD5WithRSA:
		signAlgoStr = "MD5WithRSA"
	case x509.SHA1WithRSA:
		signAlgoStr = "SHA1WithRSA"
	case x509.SHA256WithRSA:
		signAlgoStr = "SHA256WithRSA"
	case x509.SHA384WithRSA:
		signAlgoStr = "SHA384WithRSA"
	case x509.SHA512WithRSA:
		signAlgoStr = "SHA512WithRSA"
	case x509.DSAWithSHA1:
		signAlgoStr = "DSAWithSHA1"
	case x509.DSAWithSHA256:
		signAlgoStr = "DSAWithSHA256"
	case x509.ECDSAWithSHA1:
		signAlgoStr = "ECDSAWithSHA1"
	case x509.ECDSAWithSHA256:
		signAlgoStr = "ECDSAWithSHA256"
	case x509.ECDSAWithSHA384:
		signAlgoStr = "ECDSAWithSHA384"
	case x509.ECDSAWithSHA512:
		signAlgoStr = "ECDSAWithSHA512"
	default:
		signAlgoStr = "Unknown"
	}
	return signAlgoStr
}

func GetAuditType(cert *x509.Certificate) (certType int) {
	policy := false
	buss := false
	for _, ext := range cert.Extensions {
		if ext.Id.Equal(asn1.ObjectIdentifier([]int{2, 5, 29, 32})) {
			//有证书策略 ( 2.5.29.32 )
			policy = true
			break
		}
	}

	if policy {
		for _, identifier := range cert.PolicyIdentifiers {
			if identifier.Equal(asn1.ObjectIdentifier{2, 23, 140, 1, 2, 1}) {
				certType = DV
				return
			} else if identifier.Equal(asn1.ObjectIdentifier{2, 23, 140, 1, 2, 2}) {
				certType = OV
				return
			}
		}
	}

	for _, value := range cert.Subject.Names {
		if value.Type.Equal(asn1.ObjectIdentifier{2, 5, 4, 15}) {
			//获取商业类别
			buss = true
			break
		}
	}

	if cert.Subject.SerialNumber != "" && buss {
		certType = EV //EV
		return
	}

	//没有国家
	if len(cert.Subject.Country) == 0 {
		certType = DV
		return
	}

	//没有组织
	if len(cert.Subject.Organization) == 0 || cert.Subject.CommonName == cert.Subject.Organization[0] {
		certType = DV //DV
		return
	}

	certType = OV //OV
	return certType
}

// generateSubject 生成Subject信息
type Subject struct {
	CommonName         string   `json:"common_name"`
	EmailAddress       []string `json:"email_address"`
	Organization       string   `json:"organization"`
	OrganizationalUnit string   `json:"organizational_unit"`
	StreetAddress      string   `json:"street_address"`
	Locality           string   `json:"locality"`
	Province           string   `json:"province"`
	Country            string   `json:"country"`
}

func GenerateSubject(subject pkix.Name) (s *Subject) {
	s = &Subject{}
	s.CommonName = subject.CommonName
	if len(subject.Organization) != 0 {
		s.Organization = strings.Join(subject.Organization, "|")
	}
	if len(subject.OrganizationalUnit) != 0 {
		s.OrganizationalUnit = strings.Join(subject.OrganizationalUnit, "|")
	}

	if len(subject.StreetAddress) != 0 {
		s.StreetAddress = subject.StreetAddress[0]
	}

	if len(subject.Locality) != 0 {
		s.Locality = subject.Locality[0]
	}

	if len(subject.Province) != 0 {
		s.Province = subject.Province[0]
	}

	if len(subject.Country) != 0 {
		s.Country = subject.Country[0]
	}

	for _, v := range subject.Names {
		if v.Type.String() == "1.2.840.113549.1.9.1" {
			s.EmailAddress = append(s.EmailAddress, v.Value.(string))
		}
	}

	return
}

func GenerateExtKeyUsage(usage []x509.ExtKeyUsage) (str string) {
	for _, v := range usage {
		if str != "" {
			str += ","
		}
		str += ExtKeyUsageToString[v]
	}

	return
}

//区分出中间证书和根证书
func SeparateRootAndImIntermediate(certs []*x509.Certificate) (roots []*x509.Certificate, intermediates []*x509.Certificate) {
	roots = make([]*x509.Certificate, 0)
	intermediates = make([]*x509.Certificate, 0)

	for _, c := range certs {
		if certutils.CheckRoot(c) {
			roots = append(roots, c)
		} else {
			intermediates = append(intermediates, c)
		}
	}
	return
}

var ExtKeyUsageToString = map[x509.ExtKeyUsage]string{
	x509.ExtKeyUsageAny:                        "Any extended key usage",
	x509.ExtKeyUsageServerAuth:                 "Server authentication",
	x509.ExtKeyUsageClientAuth:                 "Client authentication",
	x509.ExtKeyUsageCodeSigning:                "Code signing",
	x509.ExtKeyUsageEmailProtection:            "E-mail protection",
	x509.ExtKeyUsageIPSECEndSystem:             "IP security end system",
	x509.ExtKeyUsageIPSECTunnel:                "IP security tunnel termination",
	x509.ExtKeyUsageIPSECUser:                  "IP security user",
	x509.ExtKeyUsageTimeStamping:               "Timestamping",
	x509.ExtKeyUsageOCSPSigning:                "OCSP signing",
	x509.ExtKeyUsageMicrosoftServerGatedCrypto: "Microsoft Server Gated Cryptography",
	x509.ExtKeyUsageNetscapeServerGatedCrypto:  "Netscape Server Gated Cryptography",
}

var KeyUsageToString = map[x509.KeyUsage]string{
	x509.KeyUsageDigitalSignature:  "DigitalSignature",
	x509.KeyUsageContentCommitment: "ContentCommitment",
	x509.KeyUsageKeyEncipherment:   "KeyEncipherment",
	x509.KeyUsageDataEncipherment:  "DataEncipherment",
	x509.KeyUsageKeyAgreement:      "KeyAgreement",
	x509.KeyUsageCertSign:          "CertSign",
	x509.KeyUsageCRLSign:           "CRLSign",
	x509.KeyUsageEncipherOnly:      "EncipherOnly",
	x509.KeyUsageDecipherOnly:      "DecipherOnly",
}

// 证书类型
type CertDomainType int

const (
	DomainTypeStandard   CertDomainType = 1
	DomainTypeWildCard                  = 2
	DomainTypeMutiDomain                = 4
)

func GetDomainType(cn string, sans []string) (typ CertDomainType) {
	l := len(sans)

	// 判断是否为购买WWW、通配符等，送主域名？（不算多域名，同时存在两个域名
	var giveMater bool
	if l > 1 {
		cnLow := strings.ToLower(cn)
		if strings.HasPrefix(cnLow, "www.") {
			cnLow = strings.TrimPrefix(cnLow, "www.")
		} else {
			cnLow = "www." + cnLow
		}

		contain := func(d string, list []string) bool {
			for _, v := range list {
				if d == v {
					return true
				}
			}
			return false
		}

		if contain(cnLow, sans) {
			giveMater = true
		} else {
			cnLow = strings.ToLower(cn)
			if strings.HasPrefix(cnLow, "*.") {
				cnLow = strings.TrimPrefix(cnLow, "*.")
			} else {
				cnLow = "*." + cnLow
			}
			giveMater = contain(cnLow, sans)
		}
	}

	if giveMater {
		l--
	}

	for _, dnsName := range sans {
		if strings.HasPrefix(dnsName, "*.") {
			typ = DomainTypeWildCard
		}
	}

	if l > 1 {
		typ = typ | DomainTypeMutiDomain
	} else if l == 1 && (typ != DomainTypeWildCard) {
		// 不会有 单域名通配符
		typ = DomainTypeStandard
	}

	return
}
