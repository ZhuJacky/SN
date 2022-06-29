package model

import (
	"crypto/x509"
	"math"
	"time"

	"mysslee_qcloud/common"
)

type SecureLevel uint8

const (
	ProgressErr  = "error" //错误
	ProgressDone = "done"  //完成
)

type MessageForUnmarshal struct {
	Error     string  `json:"error"`              //错误信息
	ErrorCode int     `json:"error_code"`         //错误码
	Progress  string  `json:"progress"`           //正在进行的内容
	DataMap   int     `json:"data_map,omitempty"` //已完成的数据信息
	Data      *Report `json:"data"`               //检测的数据
}

type Report struct {
	Version string                   `json:"_version"`
	Basic   *Basic                   `json:"basic"`
	Certs   *MultipleCertificateInfo `json:"certificates"`
	Ciphers *CiphersOfProtoDescribe  `json:"ciphers"`
	// Protocols []*Describe              `json:"protos"`
	Bugs *Bugs `json:"bugs"`
	// Simulate  []*SimulateResponse      `json:"simulate"`
	// Detail            *Detail                  `json:"detail"`
	// CertCompatibility []*CertPlatformTrust `json:"cert_compatibility"`
}

type Basic struct {
	// EvaluateDate     string    `json:"evaluate_date"` //评估日期
	// Duration         int       `json:"duration"`    //检测持续时间
	Domain           string    `json:"domain"`      //检测域名
	IP               string    `json:"ip"`          //检测的IP
	IPLocation       string    `json:"ip_location"` //IP地理位置
	Port             string    `json:"port"`        //端口
	Level            uint8     `json:"level"`
	LevelStr         string    `json:"level_str"`
	DemotionReason   []*Reason `json:"demotion_reason"`    //降级原因
	IsATS            bool      `json:"is_ats"`             //是否满足ATS标准
	IsPCI            bool      `json:"is_pci"`             //是否满足PCI DSS标准
	IgnoreTrustLevel string    `json:"ignore_trust_level"` //忽略证书信息问题的等级
	Title            string    `json:"title"`              //网页标题
	Server           string    `json:"server"`             // 服务器类型
	IconURL          string    `json:"icon_url"`           // favicon.ico
}

type Reason struct {
	Id       uint   `json:"id"`
	Level    string `json:"level"`
	Describe string `json:"describe"`
}

//多证书
type MultipleCertificateInfo struct {
	RSAs []*CertificateAndChain `json:"rsas"`
	ECCs []*CertificateAndChain `json:"eccs"`
}

//证书及证书链信息
type CertificateAndChain struct {
	Certs           []*x509.Certificate `json:"-"`
	LeafCertInfo    *LeafCertInfo       `json:"leaf_cert_info"`
	CertsFormServer *CertsFormServer    `json:"certs_form_server"`
	Chain           *Chain              `json:"chain"`
}

//LeafCertInfo 叶子证书信息
type LeafCertInfo struct {
	Hash               string   `json:"hash"`
	CertStatus         int      `json:"cert_status"`
	CertStatusText     string   `json:"cert_status_text"`
	CommonName         string   `json:"common_name"`         //通用名称
	PublicKeyAlgorithm string   `json:"publickey_algorithm"` //加密算法
	Issuer             string   `json:"issuer"`              //签发者信息
	SignatureAlgorithm string   `json:"signature_algorithm"` //签名算法
	Organization       string   `json:"organization"`        //组织
	OrganizationUnit   string   `json:"organization_unit"`   //组织部门
	DNSs               []string `json:"sans"`                //备用名称
	Transparency       string   `json:"transparency"`        //透明计划
	CertType           string   `json:"cert_type"`           //证书类型 1DV，2OV，3EV
	CertDomainType     int      `json:"cert_domain_type"`    //证书域名类型
	BrandName          string   `json:"brand_name"`          //证书品牌
	ValidFrom          string   `json:"valid_from"`          //开始时间
	ValidTo            string   `json:"valid_to"`            //结束时间
	IsSNI              bool     `json:"is_sni"`              //通过SNI获取
	OCSPMustStaple     bool     `json:"ocsp_must_staple"`    //ocsp必须装订
	OCSPUrl            []string `json:"ocsp_url"`            //ocps的url地址
	OCSPStaplingStatus int      `json:"ocsp_status"`         //ocsp状态(0: good ,1 revoke 2 unknown ,3 unsupport)
}

type CertsFormServer struct {
	ProvidedNumber int               `json:"provided_number"`
	Certs          []*SimpleCertInfo `json:"certs"`
}

type SimpleCertInfo struct {
	Order     int    `json:"order"`              //顺序
	Cn        string `json:"common_name"`        //通用名称
	Hash      string `json:"hash"`               //证书hash
	Pin       string `json:"pin"`                //证书公钥sha256
	ValidTo   string `json:"valid_to"`           //有效期
	IsExpired bool   `json:"is_expired"`         //过期
	KeyAlgo   string `json:"key_algo"`           //公钥算法
	SignAlgo  string `json:"sign_algo"`          //签名算法
	IssuerCn  string `json:"issuer_common_name"` //签发者CN
}
type Chain struct {
	Chain      []*OneOfChainInfo `json:"certs"`
	HasRoot    bool              `json:"has_root"`     //有根证书
	HasOtherCA bool              `json:"has_other_ca"` //有其他的CA证书
	MissCA     bool              `json:"miss_ca"`
	Id         string            `json:"id"`
}

type OneOfChainInfo struct {
	//证书链中类别
	Type string `json:"type"`
	// 缺链
	Missing bool `json:"missing"`
	// 来自服务器
	FromServer bool `json:"from_server"`
	// 来自服务器的顺序，0表示不是服务端传过来的
	CommonName string `json:"common_name"`
	//加密算法
	PublicKeyAlgorithm string `json:"publickey_algorithm"`
	//签名算法
	SignatureAlgorithm string `json:"signature_algorithm"`
	//证书指纹
	SHA1 string `json:"sha1"`
	//hpkp中的pin信息
	PIN string `json:"pin"`
	//有效期
	ExpiresIn int    `json:"expires_in"` //剩余天数
	Issuer    string `json:"issuer"`     //签发者
	BeginTime string `json:"begin_time"` //开始时间
	EndTime   string `json:"end_time"`   //结束时间
	Order     int    `json:"order"`      //证书在服务器的顺序
	IsCa      bool   `json:"is_ca"`      //是否是ca证书
}

type CiphersOfProtoDescribe struct {
	HavePrefer    bool           `json:"have_prefer"`
	SSL2Describe  *ServerCiphers `json:"ssl_2_describe"`
	SSL3Describe  *ServerCiphers `json:"ssl_3_describe"`
	TLS10Describe *ServerCiphers `json:"tls_10_describe"`
	TLS11Describe *ServerCiphers `json:"tls_11_describe"`
	TLS12Describe *ServerCiphers `json:"tls_12_describe"`
	TLS13Describe *ServerCiphers `json:"tls_13_describe"`
}

type ServerCiphers struct {
	CipherDescribes []*CipherDescribe `json:"cipher_describes"`
}

type CipherDescribe struct {
	Name     string      `json:"name"` //套件名
	Bits     uint        `json:"bits"`
	Code     string      `json:"code"`
	Secure   SecureLevel `json:"secure"`
	Describe string      `json:"describe"`
	FS       bool        `json:"fs"`
	Special  bool        `json:"special"`
}

//Describe 描述
type Describe struct {
	Key   string `json:"key"`      // 键
	Value int    `json:"value"`    // 值
	Desc  string `json:"describe"` // 描述
	Level int    `json:"level"`    // 颜色等级
}

type Bugs struct {
	Drown         *BugDescribe `json:"drown"`
	CCS           *BugDescribe `json:"ccs"`
	Heartbleed    *BugDescribe `json:"heartbleed"`
	PaddingOracle *BugDescribe `json:"padding_oracle"`
	TLSPOODLE     *BugDescribe `json:"tlspoodle"`
	FREAK         *BugDescribe `json:"freak"`
	Logjam        *BugDescribe `json:"logjam"`
	POODLE        *BugDescribe `json:"poodle"`
	CRIME         *BugDescribe `json:"crime"`
	RobotDetect   *BugDescribe `json:"robot_detect"`
}

type BugDescribe struct {
	Support int    `json:"support"`
	Danger  string `json:"danger"`
}

type CertPlatformTrust struct {
	CertType      string           `json:"cert_type"`
	CertIndex     int              `json:"cert_index"`
	CertID        string           `json:"cert_id"`
	PlatformTrust []*PlatformTrust `json:"platform_trust"`
}

type PlatformTrust struct {
	Platform string `json:"paltform"`
	Pass     bool   `json:"pass"`
	Comments string `json:"comments"`
}

type DualFullCertificateInfo struct {
	RSA *CertificateAndChain `json:"rsa"`
	ECC *CertificateAndChain `json:"ecc"`
}

//详细说明
type Detail struct {
	MailServer                  bool      `json:"mail_server"`                   //是否是邮件服务器
	SupportStarttls             bool      `json:"support_starttls"`              //使用STARTTLS命令开启tls
	DowngradeAttackPrevention   *Describe `json:"downgrade_attack_prevention"`   //降级预防
	ServerSecureRenegotiation   *Describe `json:"server_secure_renegotiation"`   //服务器安全重协商
	ClientInsecureRenegotiation *Describe `json:"client_insecure_renegotiation"` //客户端不安全重协商
	ClientSecureRenegotiation   *Describe `json:"client_secure_renegotiation"`   //客户端安全重协商
	RC4                         *Describe `json:"rc4"`                           //使用RC4加密套件
	PerfectForwardSecrecy       *Describe `json:"perfect_forward_secrecy"`       //支持正向保密
	HPKP                        *Describe `json:"hpkp"`                          //启用HPKP
	HPKPReportOnly              *Describe `json:"hpkp_report_only"`              //启用HPKP仅报告
	HSTS                        *Describe `json:"hsts"`                          //启用HSTS
	ALPN                        *Describe `json:"alpn"`                          //支持ALPN
	NPN                         *Describe `json:"npn"`                           //支持NPN
	Heartbeat                   *Describe `json:"heartbeat"`                     //使用HeartBeat扩展
	OCSP                        *Describe `json:"ocsp"`                          //支持OCPSPling
	SupportECCurves             *Describe `json:"support_ec_curves"`             //所有支持的椭圆曲线
	SSL2Compatibility           *Describe `json:"ssl2_compatibility"`            //支持23兼容包
	CAA                         *Describe `json:"caa"`                           // 支持DNS CAA记录
	SupportHTTP2                bool      `json:"support_http2"`                 //支持http2
	SupportSessionTicket        *Describe `json:"support_session_ticket"`        //支持Session Ticket方式会话复用
	SupportSessionCache         *Describe `json:"support_session_cache"`         //支持Session Cache方式会话复用
	LongHandshakeIntolerance    bool      `json:"long_handshake_intolerance"`    //长握手容忍
	TLSVersionIntolerance       *Describe `json:"tls_version_intolerance"`       //tls 版本不容忍
	IncorrectSNIAlert           bool      `json:"incorrect_sni_alert"`           //不正确sni警告
	DHPubkeyParamReuse          *Describe `json:"dh_pubkey_param_reuse"`         //dh 公钥参数重用
	ECDHPubkeyParamReuse        *Describe `json:"ecdh_pubkey_param_reuse"`       //ecdh 公钥参数重用
	UseAEADCipher               bool      `json:"use_aead_cipher"`               //使用AEAD类型的加密套件
	SupportTLS13                bool      `json:"support_tls_13"`                //支持tls13 draft 18
}

type SimulateResponse struct {
	ID                    uint16               `json:"id"` //内部id
	ClientId              int                  `json:"client_id"`
	ClientName            string               `json:"client_name"` //客户端名字和版本
	NotSupportSni         bool                 `json:"not_support_sni"`
	NotSupportPFS         bool                 `json:"not_support_pfs"`
	Code                  int                  `json:"code"`                     //检测的结果
	ProtoOrCipherMismatch bool                 `json:"proto_or_cipher_mismatch"` //协议或者加密套件不匹配
	Certificate           *SimulateCertificate `json:"certificate"`
	ServerHello           *SimulateServerHello `json:"server_hello"`
	CipherInfo            *SimulateCipher      `json:"cipher_info"`
	ErrorMsg              string               `json:"error_message"`
}

type SimulateCertificate struct {
	Hash string `json:"hash"`
	//Issuer   string `json:"-"`
	PubAlgo  string `json:"publickey_algo"`
	Strength int    `json:"strength"`
	SignAlgo string `json:"signature_algo"`
	//Trust    bool   `json:""`
	Describe string `json:"describe"`
}

type SimulateCipher struct {
	Name   string `json:"name"` //套件名称
	Code   string `json:"code"`
	IsPFS  bool   `json:"is_pfs"` //支持前向保密
	IsRc4  bool   `json:"is_rc4"`
	KxInfo string `json:"kx_info"`
}

type SimulateServerHello struct {
	IsSSL2              bool     `json:"is_ssl2"` //支持ssl2
	Version             string   `json:"version"` //支持的版本
	NegotiationProtocol []string `json:"npn"`     //协商的协议
	ALPN                string   `json:"alpn"`
}

const (
	BugUnsupported = iota
	BugSupportStrong
	BugSupportWeak
	BugUnknown
)

func DistinguishCerts(report *Report) (sniCerts, notSniCerts []*LeafCertInfo) {
	eccs := report.Certs.ECCs
	rsas := report.Certs.RSAs

	//优先ECC 证书，然后RSA证书，区分出通过sni和非sni获取的证书
	for _, ecc := range eccs {
		if ecc.LeafCertInfo.IsSNI {
			sniCerts = append(sniCerts, ecc.LeafCertInfo)
		} else {
			notSniCerts = append(notSniCerts, ecc.LeafCertInfo)
		}
	}

	for _, rsa := range rsas {
		if rsa.LeafCertInfo.IsSNI {
			sniCerts = append(sniCerts, rsa.LeafCertInfo)
		} else {
			notSniCerts = append(notSniCerts, rsa.LeafCertInfo)
		}
	}
	return
}

func GetCertValidaty(leaf *LeafCertInfo) string {
	validTo, _ := time.Parse(time.RFC3339, leaf.ValidTo)
	days := math.Ceil(validTo.UTC().Sub(time.Now().UTC()).Hours() / 24)

	if days <= 0 {
		return common.ValidityLt0
	} else if days < 30 {
		return common.ValidityLt30
	} else if days < 60 {
		return common.ValidityLt60
	} else if days < 90 {
		return common.ValidityLt90
	} else {
		return common.ValidityGte90
	}
}
