package ssl

import (
	"context"
	"crypto/x509"
	"errors"

	"mysslee_qcloud/utils"
)

type CompatibilityHandshake struct {
	SSL2ServerHandshakeMsg *SSL2ServerHandshakeMsg
	SSL3ServerHandshakeMsg *ServerHandshakeMsg
}

type SSL2ServerHandshakeMsg struct {
	Version  []byte
	CertType byte
	CertRaw  []byte
	Cipher   []byte
}

func GenerateSSL2ServerHandshakeMsg(ctx context.Context, data []byte) (*SSL2ServerHandshakeMsg, error) {
	defer utils.Recover(ctx)
	index := 0

	if len(data) <= 3 {
		return nil, errors.New("SSL2 ServerHello 数据包过短")
	}

	//暂时没有escape的情况
	//isescape:=(data[0]&0x40)!=0

	//ssl2的数据长度是前一个字节的后4位与第二个字节组成的
	a := byte(data[0] & 0x7F)  //剔除标志位 10000000
	lenByte := make([]byte, 2) //长度数据包
	lenByte[0] = a
	lenByte[1] = data[1]

	totalLen := utils.BytesToInt(lenByte) //获取总长度

	if totalLen+2 != len(data) {
		return nil, errors.New("SSL2 ServerHello 数据包不完整")
	}

	msg := &SSL2ServerHandshakeMsg{}

	index += 2 //ssl2标志位和总长度
	index += 1 //HandShake Message Type
	index += 1 //Session ID Hit

	msg.CertType = data[index]
	index += 1 //Certificate Type

	msg.Version = data[index : index+2]
	index += 2 //Version

	certLen := utils.BytesToInt(data[index : index+2])
	index += 2 //Certificate Length

	cipherLen := utils.BytesToInt(data[index : index+2])
	index += 2 //Cipher length

	//connectionIDLen:=utils.BytesToInt(data[index:index+2])
	index += 2 //

	msg.CertRaw = data[index : index+certLen]
	index += certLen //Certificate

	msg.Cipher = data[index : index+cipherLen]

	index += cipherLen
	return msg, nil
}

func GetSSL2CiphersInfo(ciphers []byte) ([]*CipherInfo, error) {
	if len(ciphers) == 0 {
		return make([]*CipherInfo, 0), nil
	}

	if len(ciphers)%3 != 0 {
		return nil, errors.New("得到的ssl2加密套件数据错误")
	}

	infos := make([]*CipherInfo, 0)
	for i := 0; i < len(ciphers); i += 3 {
		code := CipherID(uint32(ciphers[i])<<16 + uint32(ciphers[i+1])<<8 + uint32(ciphers[i+2]))
		info, exist := code.GetCipherInfo()
		if exist {
			infos = append(infos, info)
		}
	}
	return infos, nil
}

//获取所有的证书
func (s *SSL2ServerHandshakeMsg) Certs() ([]*x509.Certificate, error) {
	if len(s.CertRaw) > 0 {
		return x509.ParseCertificates(s.CertRaw)
	}
	return nil, errors.New("无法解析出证书信息")
}
