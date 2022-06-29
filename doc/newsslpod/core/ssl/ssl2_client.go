package ssl

import (
	"context"
	"errors"
	"io"
	"net"
	"time"

	"mysslee_qcloud/core/myerr"
	"mysslee_qcloud/utils"

	log "github.com/sirupsen/logrus"
)

type SSL2ClientHelloOpt struct {
	Ciphers   []CipherID
	Challenge []byte
}

func (o *SSL2ClientHelloOpt) GenerateClientHello(compatibility bool, maxVersion uint16) ([]byte, error) {
	index := 0

	if len(o.Ciphers) == 0 {
		return nil, errors.New("加密套件数量为空")
	}

	cipherLen := len(o.Ciphers) * 3 //计算ssl2的加密套件长度
	var haveChallenge bool
	var challengeLen int
	if len(o.Challenge) != 0 {
		haveChallenge = true
		challengeLen = len(o.Challenge)
	}
	totalLen := 0
	totalLen += 1         //HandshakeMessage
	totalLen += 2         //Version
	totalLen += 2         //Cipher Length
	totalLen += 2         //Session ID Length
	totalLen += 2         //Challenge length
	totalLen += cipherLen //ciphers data
	if !haveChallenge {
		totalLen += 16
	}
	totalLen += challengeLen
	//challenge data

	data := make([]byte, totalLen+2)

	//data[0] = 0x80 //ssl2 标志
	index += 1

	//Length
	data[1] = byte(totalLen & 0xFF)
	index += 1
	data[0] = 0x80 + byte((totalLen>>8)&0x0F) //标志加长度高位

	data[2] = 0x01 //ClientHello
	index += 1

	//Version
	if compatibility { //发送是否是23兼容包
		data[3] = byte(maxVersion >> 8)
		data[4] = byte(maxVersion & 0xff)
	} else {
		data[3] = 0x00
		data[4] = 0x02
	}
	index += 2

	//CipherLength
	data[5] = byte(cipherLen >> 8)
	data[6] = byte(cipherLen & 0xff)
	index += 2

	//Session ID Length
	data[7] = 0x00
	data[8] = 0x00
	index += 2

	//Challenge Length

	if haveChallenge {
		d, err := utils.IntToByteArray(len(o.Challenge))
		if err != nil {
			return nil, err
		}

		if len(d) != 2 {
			return nil, errors.New("获取长度错误")
		}

		data[9] = d[0]
		data[10] = d[1]

	} else {
		data[9] = 0x00
		data[10] = 0x10
	}

	index += 2

	//Cipher
	cipherData := ssl2CiphersToRawBytes(o.Ciphers)
	copy(data[index:index+cipherLen], cipherData)
	index += cipherLen

	//Challenge
	if !haveChallenge {
		copy(data[index:], ChallengeData)
	} else {
		copy(data[index:], o.Challenge)
	}
	return data, nil
}

func (o *SSL2ClientHelloOpt) MakeClientHello(compatibility bool, maxVersion uint16, ciphers []CipherID) ([]byte, error) {
	o.Ciphers = ciphers
	return o.GenerateClientHello(compatibility, maxVersion)

}

func (o *SSL2ClientHelloOpt) GetServerHandshakeMsg(ctx context.Context, conn net.Conn, compatibility bool, maxVersion uint16, ciphers []CipherID) (msg *CompatibilityHandshake, err error) {

	data, err := o.MakeClientHello(compatibility, maxVersion, ciphers)
	if err != nil {
		return nil, err
	}

	_, err = conn.Write(data)
	if err != nil {
		return nil, err
	}

	return GetSSL23Handshake(ctx, conn)
}

//处理23兼容包的问题
func GetSSL23Handshake(ctx context.Context, conn net.Conn) (*CompatibilityHandshake, error) {
	var ssl2Handshake *SSL2ServerHandshakeMsg
	var ssl3Handshake *ServerHandshakeMsg
	var data []byte
	var err error
	var sslType = make([]byte, 1)
	var n int

	conn.SetReadDeadline(time.Now().Add(10 * time.Second)) //设置读取死亡时间

	n, err = conn.Read(sslType)
	if n == 0 && err == io.EOF {
		return nil, myerr.ErrServerClose
	}

	if err != nil {
		return nil, err
	}
	if n == 0 {
		return nil, errors.New("ssl2握手没有获取到完整的类型")
	}

	if sslType[0] >= 80 {
		//走ssl2解析
		data, err = GetSSL2HandshakeData(conn, sslType[0])
		if err != nil {
			return nil, err
		}
		ssl2Handshake, err = GenerateSSL2ServerHandshakeMsg(ctx, data)

	} else if sslType[0] == 0x15 || sslType[0] == 0x016 {
		//走ssl3解析
		ssl3Handshake, err = GetSSL3HandshakeData(ctx, conn, sslType[0])
	} else {
		return nil, errors.New("非tls握手信息")
	}

	if err != nil {
		return nil, err
	}

	handshake := &CompatibilityHandshake{}

	handshake.SSL2ServerHandshakeMsg = ssl2Handshake
	handshake.SSL3ServerHandshakeMsg = ssl3Handshake
	return handshake, nil

}

//获取ssl2握手信息
func GetSSL2HandshakeData(conn net.Conn, typebyte byte) ([]byte, error) {
	var buf []byte
	var err error

	buf = make([]byte, 2)
	_, err = conn.Read(buf)
	if err != nil {
		return nil, errors.New("读取SSL2Handshake数据错误")
	}
	buf = append([]byte{typebyte}, buf...)

	if buf[0] < 80 || buf[2] != 4 {
		return nil, errors.New("没有读取到SSL2 ServerHello信息")
	}
	a := byte(buf[0] & 0x7F)   //剔除标志位 10000000
	lenByte := make([]byte, 2) //长度数据包
	lenByte[0] = a
	lenByte[1] = buf[1]
	totalLen := utils.BytesToInt(lenByte) //获取总长度
	leftData := make([]byte, totalLen-1)  //用于存放剩下的内容
	//循环读防止出现没有读完的情况
	total := 0
	for total < totalLen-1 {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err := conn.Read(leftData[total:])
		total += n
		if total == 0 && err == io.EOF {
			return nil, err
		}
		if err != nil {
			if err != io.EOF {
				return nil, errors.New("获取完整的SSL2 Handshake信息错误")
			}
			break
		}
		//防止意外死循环
		if n == 0 {
			return nil, myerr.ErrReadZeroData
		}
	}
	return append(buf, leftData[:]...), nil
}

//获取ssl3握手信息
func GetSSL3HandshakeData(ctx context.Context, conn net.Conn, typebyte byte) (*ServerHandshakeMsg, error) {
	conn1, conn2 := net.Pipe()
	defer conn1.Close()

	// Pipe conn
	go func() {
		defer utils.Recover(ctx)
		defer conn2.Close()
		conn1.Write([]byte{typebyte})
		var buf = make([]byte, 1024)
		for {
			conn.SetReadDeadline(time.Now().Add(10 * time.Second))
			n, err := conn.Read(buf)
			if n > 0 {
				conn1.Write(buf[:n])
			}

			if err != nil || n == 0 {
				if err == nil {
					log.Info("pipe break with n=0!!!")
				}
				break
			}
		}
	}()
	return getHandshakeMessage(ctx, conn2)
}
