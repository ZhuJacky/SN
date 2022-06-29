package core

import (
	"context"
	"crypto/x509"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	certutil "mysslee_qcloud/core/cert"
	"mysslee_qcloud/core/myconn"
	"mysslee_qcloud/core/myerr"
	"mysslee_qcloud/core/ocsp"
	"mysslee_qcloud/core/runner"
	"mysslee_qcloud/core/ssl"
	"mysslee_qcloud/dns"
	"mysslee_qcloud/utils"
	"mysslee_qcloud/utils/certutils"

	log "github.com/sirupsen/logrus"
	"github.com/tjfoc/gmsm/sm2"
)

const (
	CertTrust           = 0x0000
	CertUseWeakSignAlgo = 0x0001
	CertUntrust         = 0x0002
	CertNameUnmatch     = 0x0004
	CertInBlack         = 0x0008
	CertRevoke          = 0x0010
	CertExpired         = 0x0020
)
const (
	defaultCheck = 0
	eccCheck     = 1
	rsaCheck     = 2
)

//OutCerts 输出的
type OutCerts struct {
	IP     string // ip地址
	Port   string // 端口
	Domain string // 域名

	CN           string               // subject CommonName
	DNSS         []string             // DNSNames
	CertFromSSL2 bool                 // 从ssl2 或者ssl2 兼容包获取的证书
	SNI          bool                 // 是否使用SNI技术
	DomainInCert bool                 // 验证域名是否在证书中
	Expires      int                  // 是否过期
	TrustStatus  certutil.TrustStatus // 是否可以信任
	CertsInfo    []*certutil.CertInfo // 证书信息（从服务器端获取的证书信息）
	ChainStatus  byte                 // 证书链状态
	NotExtra     bool                 // 主要的
	Default      bool                 // 默认获取的证书

	OCSPMustStaple   bool              // OCSP必须装订
	OCSPStapling     bool              // 进行OCSP验证
	OCSPStaplingRaw  []byte            // ocspStapling原始信息
	OCSPUrl          []string          // ocsp地址
	OCSP             certutil.OCSPInfo // OCSP信息
	OCSPStaplingInfo certutil.OCSPInfo // OCSPStapling信息
	DoCheckOCSP      bool              // 已经做了ocsp检测了

	SCTRaw        []byte // sct 原始数据
	IsCTQualified int    // 是否符合CT政策

	LeafType           x509.PublicKeyAlgorithm
	ServerCertificates []*x509.Certificate // 服务器端证书

	IsSM2          bool
	SM2ServerCerts []*sm2.Certificate
	SM2LeafType    sm2.PublicKeyAlgorithm
}

//DualCertificate 双证书信息
type DualCertificate struct {
	RSA *SimpleCertificateInfo `json:"rsa"`
	ECC *SimpleCertificateInfo `json:"ecc"`
}

type DualFullCertificateInfo struct {
	RSA *OutCerts
	ECC *OutCerts
}

func GetMultipleCertInfo(ctx context.Context, checkParams *myconn.CheckParams) ([]*CertNode, error) {

	var certsNodes = make([]*CertNode, 0)
	var certsChan chan *OutCerts

	certsChan = make(chan *OutCerts, 6)
	var wg sync.WaitGroup

	//优先获取RSA证书
	checkDefualtCertInfo(ctx, checkParams, certsChan)

	wg.Add(3)

	//获取RSA证书
	go func() {
		defer utils.Recover(ctx)
		defer wg.Done()
		checkRsaCertInfo(ctx, checkParams, certsChan)

	}()

	//增加200ms 保证带ocspstapling的请求尽可能先完成
	//通过tls13获取证书
	time.Sleep(200 * time.Millisecond)
	go func() {
		defer utils.Recover(ctx)
		defer wg.Done()
		outcert, _ := GetTLS13Certificate(ctx, checkParams)
		if outcert != nil {
			certsChan <- outcert
		}

	}()

	//使用ssl2获取证书
	go func() {
		defer utils.Recover(ctx)
		defer wg.Done()
		outcert, _ := CheckSSL2CertInfo(ctx, checkParams, true)
		if outcert != nil {
			certsChan <- outcert
		}
	}()

	checkEccCertInfo(ctx, checkParams, certsChan)

	wg.Wait()
	close(certsChan)

	if len(certsChan) == 0 {
		return nil, errors.New("无法获取到证书")
	}

	for cert := range certsChan {
		if cert != nil {
			if !CertNodeIsExist(certsNodes, cert.CertsInfo[0].Sha1) {
				certsNodes = AddCertNode(ctx, certsNodes, cert)
			}
		}
	}

	return certsNodes, nil
}

//只获取tls13证书
func GetTLS13Certificate(ctx context.Context, params *myconn.CheckParams) (outCerts *OutCerts, err error) {
	config := &runner.Config{
		ServerName:       params.Domain,
		TLS13Variant:     runner.TLS13All,
		CipherSuites:     []uint16{runner.TLS_AES_128_GCM_SHA256, runner.TLS_AES_256_GCM_SHA384, runner.TLS_CHACHA20_POLY1305_SHA256},
		CurvePreferences: []runner.CurveID{runner.CurveP224, runner.CurveP256, runner.CurveP384, runner.CurveP521, runner.CurveX25519},
	}
	addr := fmt.Sprintf("%v:%v", params.Ip, params.Port)

	conn, err := myconn.NewWithContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	client := runner.Client(conn, config)
	defer client.Close()

	err = client.Handshake()
	if err != nil {
		return nil, err
	}

	status := client.ConnectionState()
	if status.Version != runner.VersionTLS13 {
		return nil, errors.New("没有使用TLS1.3进行握手")
	}
	return GenerateCertInfo(ctx, params, status.PeerCertificates, false, true, true, nil, nil)

}

//获取默认的加密套件
func checkDefualtCertInfo(ctx context.Context, params *myconn.CheckParams, certsChan chan *OutCerts) {
	outcert, _ := CheckCertInfo(ctx, params, defaultCheck, true, false)
	if outcert != nil {
		outcert.Default = true
		certsChan <- outcert
	}
}

func checkEccCertInfo(ctx context.Context, params *myconn.CheckParams, certsChan chan *OutCerts) {
	outcert, _ := CheckCertInfo(ctx, params, eccCheck, true, false)
	if outcert != nil {
		certsChan <- outcert
	}

	//通过非SNI检测ECC证书
	outcert, _ = CheckCertInfo(ctx, params, eccCheck, false, true)
	if outcert != nil {
		certsChan <- outcert
	}
}

func checkRsaCertInfo(ctx context.Context, params *myconn.CheckParams, certsChan chan *OutCerts) {
	//通过SNI检测RSA证书
	outcert, _ := CheckCertInfo(ctx, params, rsaCheck, true, false)
	if outcert != nil {
		certsChan <- outcert
	}
	//通过非SNI检测RSA证书
	outcert, _ = CheckCertInfo(ctx, params, rsaCheck, false, true)
	if outcert != nil {
		certsChan <- outcert
	}

}

type SimpleCertificateInfo struct {
	Hash       string   `json:"hash"`
	CommonName string   `json:"common_name"`
	Issuer     string   `json:"issuer"`
	SANS       []string `json:"sans"`
	Begin      string   `json:"begin"`
	End        string   `json:"end"`
	RAW        string   `json:"raw"`
}

// 判断域的CAA记录，暂时忽略.com.cn这样的二级域
func CheckCAA(ctx context.Context, domain string) (support bool, supportCA string) {
	ss := strings.Split(domain, ".")
	if len(ss) > 2 {
		domain = strings.Join(ss[1:], ".")
	}
	data, err := dns.LookupCAA(ctx, domain)
	if err != nil {
		return false, ""
	}
	if len(data) > 0 {

		for _, caa := range data {
			log.Println("caa result: ", caa)
			supportCA = supportCA + caa.Value + " , "
		}
		return true, supportCA
	}
	return false, ""
}

func GenerateCertInfo(ctx context.Context, checkParams *myconn.CheckParams, certs []*x509.Certificate, certFromSSL2, isSni, notExtra bool, ocspraw, sctraw []byte) (out *OutCerts, err error) {
	out = &OutCerts{}
	out.Domain = checkParams.Domain
	out.IP = checkParams.Ip
	out.Port = checkParams.Port
	out.CertFromSSL2 = certFromSSL2
	servercerts, result, err := certutil.GetFullChain(ctx, certs)
	if err != nil {
		return nil, err
	}
	out.ServerCertificates = servercerts
	out.CertsInfo = result.Certs
	out.TrustStatus = result.CertStatus
	out.ChainStatus = result.ChainStatus
	out.NotExtra = notExtra
	out.SCTRaw = sctraw
	out.OCSPStaplingRaw = ocspraw

	if len(certs) == 0 {
		return nil, myerr.ErrGetCertNoChain
	}
	out.CN = certs[0].Subject.CommonName
	out.DNSS = certs[0].DNSNames
	if certs[0].VerifyHostname(out.Domain) == nil {
		out.DomainInCert = true
	}
	out.LeafType = certs[0].PublicKeyAlgorithm
	out.Expires = int(certs[0].NotAfter.Sub(time.Now()).Hours() / 24)

	out.SNI = isSni
	out.OCSPUrl = certs[0].OCSPServer

	out.OCSPMustStaple = certutils.CheckOCSPMustStaple(certs[0])

	return
}

//只使用SSL2协议去握手
func CheckSSL2CertInfo(ctx context.Context, checkParams *myconn.CheckParams, notExtra bool) (outcerts *OutCerts, err error) {
	opt := &ssl.SSL2ClientHelloOpt{}

	conn, err := myconn.GetConn(ctx, checkParams)
	if err != nil {
		return nil, err
	}

	defer conn.Close()

	msg, err := opt.GetServerHandshakeMsg(ctx, conn, false, 0, ssl.GetOnlySSL2Ciphers())
	if err != nil {
		return nil, err
	}

	if msg != nil && msg.SSL2ServerHandshakeMsg != nil && len(msg.SSL2ServerHandshakeMsg.CertRaw) != 0 {
		certs, err := msg.SSL2ServerHandshakeMsg.Certs()
		if err != nil {
			return nil, err
		}
		return GenerateCertInfo(ctx, checkParams, certs, true, false, notExtra, nil, nil)
	}
	return nil, errors.New("无法获取到证书")
}

func GetServerHandshakeAutoCiphers(ctx context.Context, param *myconn.CheckParams, opt *ssl.ClientHelloOptions, ciphers, otherCiphers []ssl.CipherID) (msg *ssl.ServerHandshakeMsg, err error) {
	msg, err = ssl.GetServerHandshakeMsg(ctx, param, opt, ciphers)
	if err != nil {
		return
	}
	if msg != nil && len(msg.AlertMsg) > 0 {
		if ssl.IsFatalAlertMsg(msg.AlertMsg) && msg.Alert().Description == ssl.HandshakeFailure { //如果是握手Alert ，跟换
			msg, err = ssl.GetServerHandshakeMsg(ctx, param, opt, otherCiphers)
		}
	}
	return
}

//CheckCertInfo 核对证书信息
func CheckCertInfo(ctx context.Context, checkParams *myconn.CheckParams, checkType int, notExtra, disableSNI bool) (outCerts *OutCerts, err error) {

	opt := &ssl.ClientHelloOptions{
		ServerName:                        checkParams.Domain,
		MaxVersion:                        ssl.TLSv12,
		MinVersion:                        ssl.TLSv10,
		CustomSignatureHash:               true,
		SupportSignedCertificateTimestamp: true,
		SupportStatusRequest:              true,
	}

	if disableSNI {
		opt.DisableSni = true
	}

	var (
		msg *ssl.ServerHandshakeMsg
	)

	conn, err := myconn.GetConn(ctx, checkParams)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	switch checkType {
	case defaultCheck:
		msg, err = GetServerHandshakeAutoCiphers(ctx, checkParams, opt, ssl.PopularCiphers, ssl.DefaultCiphers)
	case eccCheck:
		msg, err = GetServerHandshakeAutoCiphers(ctx, checkParams, opt, ssl.PopularECCCiphers, ssl.ECCCiphers)
	case rsaCheck:
		msg, err = GetServerHandshakeAutoCiphers(ctx, checkParams, opt, ssl.PopularRSACiphers, ssl.RSACipher)
	}
	if err != nil {
		return nil, err
	}

	//tls10~tls12无法获取连接 ,回退到ssl3上进行连接
	if msg != nil && len(msg.AlertMsg) > 0 && ssl.IsFatalAlertMsg(msg.AlertMsg) && msg.Alert().Description == ssl.HandshakeFailure {
		opt.MinVersion = ssl.SSLv3
		opt.MaxVersion = ssl.SSLv3
		switch checkType {
		case defaultCheck:
			msg, err = GetServerHandshakeAutoCiphers(ctx, checkParams, opt, ssl.PopularCiphers, ssl.DefaultCiphers)
		case eccCheck:
			msg, err = GetServerHandshakeAutoCiphers(ctx, checkParams, opt, ssl.PopularECCCiphers, ssl.ECCCiphers)
		case rsaCheck:
			msg, err = GetServerHandshakeAutoCiphers(ctx, checkParams, opt, ssl.PopularRSACiphers, ssl.RSACipher)
		}
		if err != nil {
			return nil, err
		}
	}

	if msg != nil && len(msg.ServerHelloRaw) != 0 && len(msg.CertsRaw) != 0 {
		certs, err := msg.Certs()

		if err != nil {
			log.WithFields(log.Fields{
				"domain": checkParams.Domain,
				"port":   checkParams.Port,
				"ip":     checkParams.Ip,
				"line":   utils.ShowCallerMessage(1),
				"event":  "获取证书信息",
			}).Warnf("获取证书错误:%v", err)

			return nil, errors.New("解析Certs错误")
		}

		hello, err := msg.Hello()
		if err != nil {
			return nil, err
		}

		var sctraw []byte
		if ext, hasExt := hello.HelloExt(ctx); hasExt {
			sctraw = ext.SCTRaw
		}
		return GenerateCertInfo(ctx, checkParams, certs, false, !opt.DisableSni, notExtra, msg.OCSPStapling, sctraw)

	}

	return nil, errors.New("没有获取到握手信息")
}

func GetTrustStatus(status int) string {
	if status&CertExpired != 0 {
		return "已过期"
	} else if status&CertRevoke != 0 {
		return "已吊销"
	} else if status&CertInBlack != 0 {
		return "黑名单"
	} else if status&CertNameUnmatch != 0 {
		return "域名不匹配"
	} else if status&CertUntrust != 0 {
		return "不可信"
	} else if status&CertUseWeakSignAlgo != 0 {
		return "证书使用弱签名算法"
	} else {
		return "可信"
	}

}

//检测ocsp状态，分2种，一种通过ocspstapling获取的数据，第二种通过纯证书的方式获取
func (o *OutCerts) CheckOCSP(ctx context.Context) {
	ctxocsp, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	if len(o.OCSPStaplingRaw) > 0 {
		rocsp, err := ocsp.ParseResponse(o.OCSPStaplingRaw, nil)
		if err == nil {
			o.OCSPStapling = true
			o.OCSPStaplingInfo.NextUpdate = rocsp.NextUpdate
			o.OCSPStaplingInfo.ProducedAt = rocsp.ProducedAt
			o.OCSPStaplingInfo.RevocationReason = rocsp.RevocationReason
			o.OCSPStaplingInfo.Status = rocsp.Status
			o.OCSPStaplingInfo.RevokedAt = rocsp.RevokedAt
		}
	}

	if len(o.CertsInfo) > 1 { //叶子证书不为空
		if data := ocsp.GetOCSPdata(ctxocsp, o.CertsInfo[0].X509, o.CertsInfo[1].X509); data != nil {
			rocsp, err := ocsp.ParseResponse(*data, nil)
			if err == nil {
				o.OCSP.NextUpdate = rocsp.NextUpdate
				o.OCSP.ProducedAt = rocsp.ProducedAt
				o.OCSP.RevocationReason = rocsp.RevocationReason
				o.OCSP.Status = rocsp.Status
				o.OCSP.RevokedAt = rocsp.RevokedAt
			}
		}
	} else {
		if data := ocsp.GetOCSPdata(ctxocsp, o.CertsInfo[0].X509, nil); data != nil {
			rocsp, err := ocsp.ParseResponse(*data, nil)
			if err == nil {
				o.OCSP.NextUpdate = rocsp.NextUpdate
				o.OCSP.ProducedAt = rocsp.ProducedAt
				o.OCSP.RevocationReason = rocsp.RevocationReason
				o.OCSP.Status = rocsp.Status
				o.OCSP.RevokedAt = rocsp.RevokedAt
			}
		}
	}
	o.DoCheckOCSP = true
}

func IsTrust(out *OutCerts) int {
	var status = 0
	if len(out.CertsInfo) > 0 {
		cert := out.CertsInfo[0]
		if out.TrustStatus == certutil.Untrusted { //不可信
			status |= CertUntrust
		}

		if out.TrustStatus == certutil.BlackList { //黑名单
			status |= CertInBlack
		}

		if !out.DomainInCert { //域名不匹配
			//域名不在证书中
			status |= CertNameUnmatch
		}

		if certutils.IsExpired(cert.X509) { // 过期
			status |= CertExpired
		}

		if out.OCSP.Status == ocsp.Revoked || out.OCSPStaplingInfo.Status == ocsp.Revoked { //通过使用ocsp 或者ocsp装订 得到吊销
			status |= CertRevoke
		}

		if cert.SignAlgo == x509.MD2WithRSA || cert.SignAlgo == x509.MD5WithRSA || cert.SignAlgo == x509.SHA1WithRSA || cert.SignAlgo == x509.ECDSAWithSHA1 { //签名不安全
			status |= CertUseWeakSignAlgo
		}
	}
	return status
}

func IsCertTrust(out *OutCerts) bool {
	return IsTrust(out) == CertTrust
}

func ConvertPKIXNameArrayToStr(str []string) string {
	if len(str) > 0 {
		return str[0]
	}
	return ""
}
