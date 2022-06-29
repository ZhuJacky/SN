//Package core 服务器端返回的数据处理
package ssl

import (
	"context"
	"crypto/x509"
	"fmt"
	"io"
	"net"
	"time"

	"mysslee_qcloud/core/cert"
	"mysslee_qcloud/utils"

	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	recordHeaderLen int = 5

	recordChangeCipherSpec byte = 0x14
	recordAlert            byte = 0x15
	recordHandshake        byte = 0x16

	typeServerHello        byte = 0x02
	typeCertificate        byte = 0x0b
	typeServerKeyExchange  byte = 0x0c
	typeCertificateRequest byte = 0x0d
	typeServerHelloDone    byte = 0x0e
	typeCertificateStatus  byte = 0x16
)

//ServerHandshakeMsg 服务器握手的原始信息
type ServerHandshakeMsg struct {
	//握手
	Version              []byte
	CertsRaw             []byte
	ServerHelloRaw       []byte
	ServerKeyExchangeRaw []byte
	AlertMsg             []byte
	OCSPStapling         []byte
	FatalAlert           bool
	ServerHelloDone      bool
}

// 警告信息
type AlertMsg struct {
	Level       byte //等级
	Description int  //alert类型
}

//ServerHelloMsg 服务器hello的信息
type ServerHelloMsg struct {
	raw []byte
	//服务端握手
	Version uint16 //版本
	Session []byte //Session
	//Cipher      []byte
	Cipher      CipherID //使用的加密算法
	Compression byte     //压缩方法
	Extensions  []byte   //扩展信息
}

//ServerHelloExtMsg ServerHello的额外信息
type ServerHelloExtMsg struct {
	//握手
	Sni                 bool   //是否支持sni
	SecureRenegotiation bool   //安全的重商定
	SessionTicket       bool   //是否支持SessionTicket
	OCSPStapling        bool   //是否支持装订
	SCTRaw              []byte // 证书透明数据signed_certificate_timestamp
	EcPointFormatsRaw   []byte
	//ALPNRaw             []byte
	ALPNRaw []byte //ALPN信息
	//	NPNRaw     []byte
	NPN       []string //下一代协议
	Heartbeat bool     //心跳包

	SupportedVersionRaw []byte
}

//ServerKeyExchangeMsg Server秘钥交换的信息 dh,ecdh
type ServerKeyExchangeMsg struct {

	/*RSA_EXPORT*/
	ModulusLen  int
	Modulus     []byte
	ExponentLen int
	Exponent    []byte

	/*Diffie-hellman Server Params*/
	PLen int //
	P    []byte
	GLen int //
	G    []byte

	/*EC Diffie-Hellman*/
	CurveType  byte           //椭圆类型
	NamedCurve SupportedGroup //椭圆名称
	HashString string         //使用的摘要方式

	/*通用信息*/
	PubkeyLen int // 公钥长度
	Pubkey    []byte
	//SignatureAlgorithm []byte
	//SignatureLen       int
	//Signature          []byte
}

//HelloExt serverHelloExt信息
func (a *ServerHelloMsg) HelloExt(ctx context.Context) (msg *ServerHelloExtMsg, err bool) {
	if len(a.Extensions) > 0 {
		var data = a.Extensions
		var datal, i, n int
		msg = &ServerHelloExtMsg{}
		datal = len(data)
		//fmt.Printf("%p\n", &msg)
		for i < datal {
			switch utils.BytesToInt(data[i : i+2]) {
			case 0: //Extension:server_nameA
				i += 2 + 2                           //type + length
				i += utils.BytesToInt(data[i-2 : i]) //data
				msg.Sni = true
			case 0x05: //Extension status_request
				i += 2 + 2                           //type + length
				i += utils.BytesToInt(data[i-2 : i]) //data
				msg.OCSPStapling = true
			case 0x0a: //supported_groups (renamed from "elliptic_curves")
				i += 2 + 2
				i += utils.BytesToInt(data[i-2 : i])
			case 0x0B: //Extension:ec_point_formats
				i += 2 + 2                          //type + length
				n = utils.BytesToInt(data[i-2 : i]) //数据长度
				msg.EcPointFormatsRaw = data[i : i+n]
				i += n //数据
			case 0x000f: //Extension heartbeat
				msg.Heartbeat = true
				i += 2
				n := utils.BytesToInt(data[i : i+2])
				i += 2
				i += n
			case 0x10: //Extension application_layer_protocol_negotiation
				i += 2
				n = utils.BytesToInt(data[i : i+2])
				i += 2
				//msg.ALPNRaw = data[i : i+n2]
				msg.ALPNRaw = data[i+3 : i+n]
				i += n
			case 0x0012: //signed_certificate_timestamp
				i += 2
				n = utils.BytesToInt(data[i : i+2])
				msg.SCTRaw = data[i : i+n+2] // ct tls Unmarshal 需要长度
				i += 2
				i += n
			case 0x0017: //Extended Master SecretA
				i += 2
				n := utils.BytesToInt(data[i : i+2])
				i += 2
				i += n
			case 0x0023: //Extension SessionTicket TLS
				i += 2 + 2                           //type + length
				i += utils.BytesToInt(data[i-2 : i]) //data
				msg.SessionTicket = true

				// 0x0028 draft18
				// 0x0033 draft26 27 28
			case 0x0028, 0x0033: //key_share

			case 0x3374: //Extension NPN
				i += 2 //type
				n = utils.BytesToInt(data[i : i+2])
				i += 2 //len

				//msg.NPNRaw = data[i : i+n]
				offset := 0
				for offset < n {
					j := int(data[i+offset])
					offset++ //长度
					msg.NPN = append(msg.NPN, string(data[i+offset:i+offset+j]))
					offset += j
				}
				i += n
			case 0x7550: //Extension channel_id
				i += 2 //type
				n = utils.BytesToInt(data[i : i+2])
				i += n
			case 0xFF01: //Extension renegotiation_info
				i += 2 + 2                           //type + length
				i += utils.BytesToInt(data[i-2 : i]) //data
				msg.SecureRenegotiation = true

			case 0x002b: // TLSv1.3 SupportVersionExt
				i += 2
				supportVersionLen := utils.BytesToInt(data[i : i+2])
				i += 2
				msg.SupportedVersionRaw = data[i : i+supportVersionLen]
				i += supportVersionLen
			default:
				log.WithFields(log.Fields{
					"req": utils.GetReqInfoFromContext(ctx),
				}).Warnf("解析ServerHello扩展时，有未知或未处理的扩展信息:%x", utils.BytesToInt(data[i:i+2]))
				i += 2
				n = utils.BytesToInt(data[i : i+2])
				i += 2 + n
			}
		}
		return msg, true
	} else {
		return nil, false
	}
}

func (a *ServerHandshakeMsg) TLS13Hello() (msg *ServerHelloMsg, err error) { //TLS13的ServerHello没有Session 处理和TLS12以下的不同

	if len(a.ServerHelloRaw) < 38 {
		return nil, errors.New("ServerHello的数据长度不满足最低要求")
	}

	raw := a.ServerHelloRaw
	i := 0
	msg = &ServerHelloMsg{}

	// Handshake Type: Server Hello (2)
	serverHelloTypeLen := 1
	i += serverHelloTypeLen

	// Length: 86
	lengthLen := 3
	length := utils.BytesToInt(raw[i : i+lengthLen])
	i += lengthLen

	if len(raw) != length+i {
		return nil, errors.New("ServerHello的数据获取错误")
	}

	// Version: TLS 1.2 (0x0303)
	versionLen := 2
	msg.Version, err = utils.BytesToUint16(raw[i : i+versionLen])
	if err != nil {
		return nil, err
	}
	i += versionLen

	// Random: 652fc1859de6bf427649d9fe5bc9c04b41cdfb0702561d8f...
	randomLen := 32
	i += randomLen

	// Session ID Length: 0
	sessionIdLengthLen := 1
	i += sessionIdLengthLen

	// Cipher Suite: TLS_AES_128_GCM_SHA256 (0x1301)
	cipherSuiteLen := 2
	d, err := utils.BytesToUint16(raw[i : i+cipherSuiteLen])
	if err != nil {
		return nil, err
	}
	msg.Cipher = CipherID(d)
	i += cipherSuiteLen

	// Compression Method: null (0)
	compressionMethodLen := 1
	i += compressionMethodLen

	// https://git.trustasia.cn/rd/sslchecker/issues/13
	// 服务器返回的server hello 中未带扩展处理
	if len(raw) == i {
		return msg, nil
	}

	// Extensions Length: 46
	extensionLengthLen := 2
	extensionLength := utils.BytesToInt(raw[i : i+extensionLengthLen])
	i += extensionLengthLen

	if i+extensionLength > length+4 { //预防扩展长度大于总长度
		return nil, errors.New("无法获取完整的ServerHello Extension扩展")
	}
	msg.Extensions = raw[i : i+extensionLength]

	return msg, nil
}

//Hello 获取Hello信息
func (a *ServerHandshakeMsg) Hello() (msg *ServerHelloMsg, err error) {

	//fmt.Printf("%p\n", &a)
	if len(a.ServerHelloRaw) <= 34 { //version +random
		return nil, errors.New("ServerHello的数据长度不满足最低要求")
	}

	var data = a.ServerHelloRaw

	var i, n int
	msg = &ServerHelloMsg{}
	datal := len(data)
	version, err := utils.BytesToUint16(data[:2])
	if err != nil {
		return nil, err
	}
	msg.Version = version
	i = 34

	n = int(data[i])   //获取sessionid长度
	if datal < i+n+1 { //version+random+sessionLen+sesssion
		return nil, errors.New("无法获取session的长度")
	}
	i++
	msg.Session = data[i : i+n]
	i += n

	if datal < i+2 {
		return nil, errors.New("无法获取到服务端响应的加密套件信息")
	}
	msg.Cipher = CipherID(uint16(data[i])<<8 | uint16(data[i+1]))
	i += 2

	if datal < i+1 {
		return nil, errors.New("无法获取的服务器响应的支持的压缩")
	}

	msg.Compression = data[i] //那都压缩方法
	i++

	//没有把单个扩展分开
	if i < datal {
		el := int(data[i])
		el <<= 8
		i++
		el += int(data[i])
		i++
		if el+i > datal {
			return nil, errors.New("serverHello extension 解析错误")
		}

		msg.Extensions = data[i : i+el]
	}
	return msg, nil

}

// Certs 获取证书信息
func (a *ServerHandshakeMsg) Certs() ([]*x509.Certificate, error) {
	//fmt.Println(&a)
	if len(a.CertsRaw) > 0 {
		return cert.CertsRawParseCerts(a.CertsRaw)

	}
	return nil, errors.New("无法解析出证书信息")
}

func (s *ServerHandshakeMsg) Alert() *AlertMsg {
	// fmt.Printf("%v", s.AlertMsg)
	msg := &AlertMsg{}
	msg.Level = s.AlertMsg[0]
	msg.Description = int(s.AlertMsg[1])
	return msg
}

//解析tls1.3的ServerHello信息
func (s *ServerHandshakeMsg) TLS13ServerHelloDecode(ctx context.Context, conn net.Conn) error {
	var header = make([]byte, 5)
	var total = 0

	//先读取header的前5个字段
	for total < 5 {
		conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		n, err := conn.Read(header[total:])

		if n > 0 {
			total += n
		}

		if err != nil {
			return nil
		}
		if n == 0 {
			return errors.New("获取握手错误，读取到非法的长度0！")
		}
	}

	recordType, recordLen, err := s.ServerHandshakeHeaderDecode(ctx, header)
	if err != nil {
		return err
	}
	if recordType != recordHandshake {
		return errors.New("没有获取期待的record类型")
	}

	var record = make([]byte, recordLen) //构建告知的长度

	total = 0

	for total < recordLen {
		conn.SetReadDeadline(time.Now().Add(15 * time.Second))
		n, err := conn.Read(record[total:])
		if n > 0 {
			total += n
		}

		if err != nil {
			if err != io.EOF {
				return err
			}
		}

		if n == 0 {
			return errors.New("获取握手错误，读取到非法的长度0！")
		}
	}

	if record[0] != typeServerHello {
		return errors.New("没有获取到期待的Handshake Protocol类型")
	}

	s.ServerHelloRaw = record
	//直接解析ServerHello
	return nil
}

//解析头部信息
func (s *ServerHandshakeMsg) ServerHandshakeHeaderDecode(ctx context.Context, data []byte) (recordType byte, recordLen int, err error) {
	if len(data) != recordHeaderLen {
		return 0, 0, errors.New("解析握手头部信息需要5个字节")
	}

	if len(s.Version) == 0 {
		s.Version = data[1:3]
	}

	recordLen = utils.BytesToInt(data[3:5])
	if recordLen == 0 {
		return 0, 0, errors.New("错误的SSL Record长度！")
	}
	return data[0], recordLen, nil

}

func (s *ServerHandshakeMsg) ServerHandshakeDecode(ctx context.Context, recordType byte, recordLen int, data []byte) (split bool, needLen int, err error) {
	//非法的数据导致解码越界，产生解码panic，产生误报
	defer func() {
		e := recover()
		if e != nil {
			log.WithFields(log.Fields{
				"req": utils.GetReqInfoFromContext(ctx),
			})
		}
	}()

	switch recordType {
	case recordChangeCipherSpec:
		//TODO 进行ChangeCipherSpec 的解析
	case recordAlert:
		s.AlertMsg = data
		if IsFatalAlertMsg(s.AlertMsg) {
			s.FatalAlert = true
			return false, 0, err
		}
	case recordHandshake:
		dataLen := utils.BytesToInt(data[1:4])
		if recordLen-4 < dataLen { //证明出现了分包的情况
			return true, dataLen + 4, nil
		}
		var p, n int
		for p < recordLen {
			switch data[p] {
			case typeServerHello:
				n = utils.BytesToInt(data[p+1 : p+4])
				s.ServerHelloRaw = data[p+4 : p+4+n]
				p += 4 + n
			case typeCertificate:
				n = utils.BytesToInt(data[p+1 : p+4])
				s.CertsRaw = data[p+4+3 : p+4+n]
				p += 4 + n
			case typeServerKeyExchange:
				n = utils.BytesToInt(data[p+1 : p+4])
				s.ServerKeyExchangeRaw = data[p+4 : p+4+n]
				p += 4 + n
			case typeCertificateRequest:
				n = utils.BytesToInt(data[p+1 : p+4])
				p += n + 4
			case typeServerHelloDone:
				s.ServerHelloDone = true
				return false, 0, nil
			case typeCertificateStatus:
				n = utils.BytesToInt(data[p+1 : p+4])
				s.OCSPStapling = data[p+4+4 : p+4+n]
				p += 4 + n
			default:
				n = utils.BytesToInt(data[p+1 : p+4])
				log.WithFields(log.Fields{
					"handshakeType": fmt.Sprintf("%X", data[p]),
					"req":           utils.GetReqInfoFromContext(ctx),
				}).Infof("尚未支持的handshakeType")
				p += n + 4
			}
		}
	default:
		return false, 0, errors.New("尚未支持的recordType")
	}
	return false, 0, nil
}

//ServerKeyExchange 生成秘钥交换信息
func (a *ServerHandshakeMsg) ServerKeyExchange(ctx context.Context) (msg *ServerKeyExchangeMsg, err error) {

	var data = a.ServerKeyExchangeRaw
	msg = &ServerKeyExchangeMsg{}

	if len(data) == 0 {
		return msg, nil
	}

	hello, err := a.Hello()
	if err != nil {
		return nil, err
	}

	var cipherRaw = hello.Cipher

	c, exist := cipherRaw.GetCipherInfo()
	if !exist {
		log.WithFields(log.Fields{
			"req": utils.GetReqInfoFromContext(ctx),
		}).Warnf("在解析ServerkeyExchange 解析了不能识别的加密套件：0x%X", cipherRaw)
		return nil, errors.New("不能识别的加密套件")
	}

	switch c.Kx {
	case KxRSA:
		//RSA 没有Server Key Exchange 客户端生成预主密钥，使用客户端公钥对其加密，将其包含在ClientKeyExchange消息中
		// 只有出口限制的带了
		if !c.Export {
			break
		}
		var i int

		rawLen := len(data)

		//获取Modulus len
		msg.ModulusLen = utils.BytesToInt(data[i : i+2])
		i = i + 2

		if len(data) < i+msg.ModulusLen {
			return nil, errors.New("获取KeyExchange信息错误(ModulusLen)")
		}
		msg.Modulus = data[i : i+msg.ModulusLen] //取出Modulus的数据
		i += msg.ModulusLen

		if i+1 < rawLen { //得到Exponent
			msg.ExponentLen = utils.BytesToInt(data[i : i+2])
			i += 2
			if len(data) < i+msg.ExponentLen {
				return nil, errors.New("获取KeyExchange信息错误(ExponentLen)")
			}
			msg.Exponent = data[i : i+msg.ExponentLen]
			i += msg.ExponentLen
		}

		//if i+1 < rawLen {
		//	msg.PubkeyLen = utils.BytesToInt(data[i : i+2])
		//	i += 2
		//	msg.Pubkey = data[i : i+msg.PubkeyLen]
		//
		//}

		//if i+1 < rawLen { //得到Signature
		//
		//	if !strings.Contains(c.Enc, "RC") {
		//		msg.SignatureAlgorithm = data[i: i+2]
		//		i += 2
		//	}
		//	msg.SignatureLen = utils.BytesToInt(data[i: i+2])
		//	i += 2
		//	msg.Signature = data[i: i+msg.SignatureLen]
		//}

	case KxDH:
		var i int
		msg.PLen = utils.BytesToInt(data[i : i+2])
		i += 2
		if len(data) < i+msg.PLen {
			return nil, errors.New("获取KeyExchange信息错误(PLen)")
		}
		msg.P = data[i : i+msg.PLen]
		i += msg.PLen

		msg.GLen = utils.BytesToInt(data[i : i+2])
		i += 2
		if len(data) < i+msg.GLen {
			return nil, errors.New("获取KeyExchange信息错误(GLen)")
		}
		msg.G = data[i : i+msg.GLen]
		i += msg.GLen

		msg.PubkeyLen = utils.BytesToInt(data[i : i+2])
		i += 2
		if len(data) < i+msg.PubkeyLen {
			return nil, errors.New("获取KeyExchange信息错误(DH PubkeyLen)")
		}
		msg.Pubkey = data[i : i+msg.PubkeyLen]
		i += msg.PubkeyLen

		//if i+1 < rawLen {
		//	if !strings.Contains(c.Enc, "RC") {
		//		msg.SignatureAlgorithm = data[i: i+2]
		//		i += 2
		//	}
		//
		//	msg.SignatureLen = utils.BytesToInt(data[i: i+2])
		//	i += 2
		//	msg.Signature = data[i: i+msg.SignatureLen]
		//}

	case KxECDH:
		var i int
		msg.CurveType = data[i]
		i += 1 //椭圆类型
		msg.NamedCurve = SupportedGroup(data[i])<<8 | SupportedGroup(data[i+1])
		i += 2 //公钥长度
		msg.PubkeyLen = int(data[i])
		i += 1 //公钥内容
		if len(data) < i+msg.PubkeyLen {
			return nil, errors.New("获取KeyExchange信息错误(ECDH PubkeyLen)")
		}
		msg.Pubkey = data[i : i+msg.PubkeyLen]
		i += msg.PubkeyLen //Signature 算法

		//if i+1 < rawLen {
		//	if !strings.Contains(c.Enc, "RC") {
		//		msg.SignatureAlgorithm = data[i: i+2]
		//		i += 2 //Signature len
		//	}
		//	msg.SignatureLen = utils.BytesToInt(data[i: i+2])
		//	i += 2 //Signature
		//	msg.Signature = data[i: i+msg.SignatureLen]
		//}

	default:
		log.WithFields(log.Fields{
			"req": utils.GetReqInfoFromContext(ctx),
		}).Warnf("未知的密钥解析，选择的加密套件是：%v", c.Name)
	}
	return msg, nil

}
