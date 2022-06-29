//Package core   ssl层只做发送数据包和解析数据包，不进行通信连接
package ssl

import (
	"context"
	"encoding/binary"
	"errors"
	"io"
	"math/rand"
	"net"
	"strings"
	"time"

	"mysslee_qcloud/core/myconn"
	"mysslee_qcloud/core/myerr"
	"mysslee_qcloud/utils"
)

//ClientHelloOptions 客户端hello的一些配置
type ClientHelloOptions struct {
	MinVersion                        uint16            //最小版本
	MaxVersion                        uint16            //最大版本
	SupportCompression                bool              //支持压缩
	SessionId                         []byte            //会话复用的ID
	DisableSni                        bool              //支持SNI
	ServerName                        string            //SNI中填写的服务器名
	SupportStatusRequest              bool              //支持ocsp扩展
	CustomSignatureHash               bool              //支持Signature_hash扩展
	CustomSignatureHashData           []byte            //signature_hash扩展中的内容 （可选）
	CustomSupportedGroup              bool              //支持elliptic_curve扩展
	CustomSupportedGroupData          []byte            //elliptic_curve中的内容
	CustomECPointFormat               bool              //支持ec_point_format扩展
	CustomECPointFormatData           []byte            //ec_point_format中的内容
	SupportSignedCertificateTimestamp bool              //支持signed_certificate_timestamp扩展
	Type                              myconn.ServerType //类型
}

func (opt *ClientHelloOptions) GetDataLength(ciphers []CipherID) int {
	return len(opt.MakeClientHello(ciphers))
}

//MakeClientHelloFromOpt 创建ClientHello数据
func (opt *ClientHelloOptions) MakeClientHello(ciphers []CipherID) []byte {
	var (
		serverNameLen, serverNameListLen, extServerNameLen, extLen, compressionLen int
	)
	if opt.MinVersion == 0 {
		opt.MinVersion = SSLv3
	}
	if opt.MaxVersion == 0 {
		opt.MaxVersion = SSLv3
	}

	ciphersRaw := tlsCiphersToRawBytes(ciphers)
	ciphersLen := len(ciphersRaw)

	sessionIDLen := len(opt.SessionId)
	extLen = 0

	if opt.SupportCompression {
		compressionLen = 2
	} else {
		compressionLen = 1
	}

	// 支持SNI
	if !opt.DisableSni {
		// Extension: server_name (len=29)
		//     Type: server_name (0)
		//     Length: 29
		//     Server Name Indication extension
		//         Server Name list length: 27
		//         Server Name Type: host_name (0)
		//         Server Name length: 24
		//         Server Name: tls13.crypto.mozilla.org

		serverNameLen = len(opt.ServerName)
		serverNameListLen = serverNameLen + 3 //type:host_name = 0x00 占一字节
		extServerNameLen = serverNameListLen + 2
		extLen += extServerNameLen + 4 // type:server_name = 0x00,0x00    ext_server_name_len长度位位占两字节      共4字节
	}

	if opt.SupportStatusRequest {
		extLen += len(OCSPStaplingRaw)
	}

	if opt.MaxVersion == TLSv12 { //https://tools.ietf.org/html/rfc5246#section-7.4.1.4.1
		opt.CustomSignatureHash = true
	}
	if ciphersHaveECC(ciphers) { //
		opt.CustomECPointFormat = true
		opt.CustomSupportedGroup = true
	}

	if opt.CustomSignatureHash {
		if len(opt.CustomSignatureHashData) == 0 {
			opt.CustomSignatureHashData = GetDefaultSignatureHashAlgorithms()
		}
		extLen += len(opt.CustomSignatureHashData) + 6
	}

	if opt.CustomSupportedGroup {
		if len(opt.CustomSupportedGroupData) == 0 {
			opt.CustomSupportedGroupData = GetEllipticCurvesData(GetTLSSupportGroup())
		}
		extLen += len(opt.CustomSupportedGroupData) + 6
	}

	if opt.CustomECPointFormat {
		if len(opt.CustomECPointFormatData) == 0 {
			opt.CustomECPointFormatData = DefaultECPointFormat
		}
		extLen += len(opt.CustomECPointFormatData) + 5
	}

	if opt.SupportSignedCertificateTimestamp {
		extLen += len(SCTRaw)
	}

	if extLen != 0 {
		extLen = extLen + 2 //扩展长度位占两字节
	}

	var clientHelloHex []byte
	var helloLen int

	helloLen = ciphersLen + sessionIDLen + compressionLen + 0x2F
	clientHelloHex = make([]byte, helloLen+extLen)
	clientHelloHex[0] = 0x16 //Handshake

	//MinVersion
	clientHelloHex[1] = byte(opt.MinVersion >> 8)
	clientHelloHex[2] = byte(opt.MinVersion & 0xff)

	//packageLen
	clientHelloHex[3] = byte((0x2A + ciphersLen + sessionIDLen + compressionLen + extLen) >> 8)
	clientHelloHex[4] = byte((0x2A + ciphersLen + sessionIDLen + compressionLen + extLen) & 0xFF)

	clientHelloHex[5] = 0x01 //ClientHello 标识
	//length
	clientHelloHex[6] = byte((0x26 + ciphersLen + sessionIDLen + compressionLen + extLen) >> 16)
	clientHelloHex[7] = byte((0x26 + ciphersLen + sessionIDLen + compressionLen + extLen) >> 8)
	clientHelloHex[8] = byte((0x26 + ciphersLen + sessionIDLen + compressionLen + extLen) & 0xFF)

	//maxVersion
	clientHelloHex[9] = byte(opt.MaxVersion >> 8)
	clientHelloHex[10] = byte(opt.MaxVersion & 0xff)

	//Randown
	fillRandom(clientHelloHex[11:43])

	clientHelloHex[43] = byte(sessionIDLen)
	//SessionID
	if (sessionIDLen) != 0 {
		copy(clientHelloHex[44:44+sessionIDLen], opt.SessionId)
	}

	//CipherLen
	clientHelloHex[44+byte(sessionIDLen)] = byte(ciphersLen >> 8)
	clientHelloHex[45+byte(sessionIDLen)] = byte(ciphersLen & 0xff)

	//Ciphers
	copy(clientHelloHex[46+byte(sessionIDLen):], ciphersRaw)

	if !opt.SupportCompression { //没有开启压缩
		//压缩
		clientHelloHex[helloLen-2] = 0x01
		clientHelloHex[helloLen-1] = 0x00
	} else { //开启压缩
		clientHelloHex[helloLen-3] = 0x02
		clientHelloHex[helloLen-2] = 0x01
		clientHelloHex[helloLen-1] = 0x00
	}
	if extLen != 0 {
		//因为前面的ext_len算上了自己长度位占的两字节，所以在填充ext_len位的时候要减2
		clientHelloHex[helloLen] = byte((extLen - 2) >> 8)
		clientHelloHex[helloLen+1] = byte((extLen - 2) & 0xff)
		offset := 2

		if !opt.DisableSni {
			// type == 00
			clientHelloHex[helloLen+offset] = 0x00
			clientHelloHex[helloLen+offset+1] = 0x00
			clientHelloHex[helloLen+offset+2] = byte(extServerNameLen >> 8)
			clientHelloHex[helloLen+offset+3] = byte(extServerNameLen & 0xff)
			clientHelloHex[helloLen+offset+4] = byte(serverNameListLen >> 8)
			clientHelloHex[helloLen+offset+5] = byte(serverNameListLen & 0xff)
			clientHelloHex[helloLen+offset+6] = 0x00
			clientHelloHex[helloLen+offset+7] = byte(serverNameLen >> 8)
			clientHelloHex[helloLen+offset+8] = byte(serverNameLen & 0xff)
			copy(clientHelloHex[helloLen+offset+9:serverNameLen+helloLen+offset+9], []byte(opt.ServerName))
			offset += extServerNameLen + 4
		}

		if opt.CustomECPointFormat {
			length := len(opt.CustomECPointFormatData)
			elliptic := make([]byte, length+6)
			elliptic[0] = 0x00
			elliptic[1] = 0x0b // 0x00,0x0a type

			elliptic[4] = byte(length)

			result, _ := utils.IntToByteArray(length + 1)
			elliptic[2] = result[0]
			elliptic[3] = result[1] //长度

			copy(elliptic[5:], opt.CustomECPointFormatData) //复制椭圆信息

			copy(clientHelloHex[helloLen+offset:], elliptic[:])
			offset += length + 5
		}

		if opt.CustomSupportedGroup {
			length := len(opt.CustomSupportedGroupData)
			elliptic := make([]byte, length+6)
			elliptic[0] = 0x00
			elliptic[1] = 0x0a // 0x00,0x0a type

			result, _ := utils.IntToByteArray(length)
			elliptic[4] = result[0]
			elliptic[5] = result[1] //椭圆长度

			result, _ = utils.IntToByteArray(length + 2)
			elliptic[2] = result[0]
			elliptic[3] = result[1] //长度

			copy(elliptic[6:], opt.CustomSupportedGroupData) //复制椭圆信息

			copy(clientHelloHex[helloLen+offset:], elliptic[:])
			offset += length + 6
		}

		if opt.SupportStatusRequest {
			copy(clientHelloHex[helloLen+offset:], OCSPStaplingRaw[:])
			offset += len(OCSPStaplingRaw)
		}

		if opt.CustomSignatureHash {
			//添加SignatureHash
			length := len(opt.CustomSignatureHashData)
			algo := make([]byte, length+6)
			algo[0] = 0x00
			algo[1] = 0x0d
			result, _ := utils.IntToByteArray(length)
			algo[4] = result[0]
			algo[5] = result[1]
			result, _ = utils.IntToByteArray(length + 2)
			algo[2] = result[0]
			algo[3] = result[1]
			copy(algo[6:], opt.CustomSignatureHashData)
			copy(clientHelloHex[helloLen+offset:], algo[:])
			offset += length + 6
		}

		if opt.SupportSignedCertificateTimestamp {
			copy(clientHelloHex[helloLen+offset:], SCTRaw[:])
			offset += len(SCTRaw)
		}
	}

	return clientHelloHex
}

//获取server握手信息
func GetServerHandshakeMsg(ctx context.Context, param *myconn.CheckParams, opt *ClientHelloOptions, ciphers []CipherID) (msg *ServerHandshakeMsg, err error) {

	conn, err := myconn.GetConn(ctx, param)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	return GetServerHandShakeMsgWithConn(ctx, conn, opt, ciphers)
}

//通过已经建立完的链接发送数据
func GetServerHandShakeMsgWithConn(ctx context.Context, conn net.Conn, opt *ClientHelloOptions, ciphers []CipherID) (msg *ServerHandshakeMsg, err error) {
	if len(ciphers) == 0 {
		return nil, errors.New("没有提供加密套件")
	}
	_, err = conn.Write(opt.MakeClientHello(ciphers))
	if err != nil {
		return nil, err
	}

	msg, err = getHandshakeMessage(ctx, conn)
	if err != nil {
		return nil, err
	}
	if msg == nil {
		return nil, errors.New("无效的Server Handshake Msg信息")
	}
	return msg, err
}

func GetServerHelloMsgByData(data []byte) (msg *ServerHandshakeMsg, finish bool, err error) {

	if len(data) < 5 {
		return nil, false, nil
	}

	if data[0] != 0x16 {
		return nil, false, errors.New("数据包不是Handshake 类型")
	}

	recordLen := utils.BytesToInt(data[3:5])
	if len(data) < recordLen+recordHeaderLen {
		return nil, false, nil
	}

	if data[5] != typeServerHello {
		return nil, false, errors.New("不是ServerHello数据包")
	}

	msg = &ServerHandshakeMsg{}
	msg.ServerHelloRaw = data[4+recordHeaderLen : recordLen+recordHeaderLen]

	return msg, true, nil

}

func getHandshakeMessage(ctx context.Context, conn net.Conn) (msg *ServerHandshakeMsg, err error) {
	msg = &ServerHandshakeMsg{}

	var split bool
	var buf = make([]byte, 1024)
	var totalBuf = make([]byte, 0)
	var n, count, total, current, newOffset, recordLen, needLen int
	var recordType byte

	for {
		conn.SetReadDeadline(time.Now().Add(10 * time.Second))
		n, err = conn.Read(buf)
		if n > 0 {
			totalBuf = append(totalBuf, buf[:n]...)
			total += n
			newOffset = current

		Again:
			if total < newOffset+recordHeaderLen {
				continue
			}

			recordType, recordLen, err = msg.ServerHandshakeHeaderDecode(ctx, totalBuf[newOffset:newOffset+recordHeaderLen])
			if err != nil {
				return nil, err
			}

			newOffset += recordHeaderLen
			if total < newOffset+recordLen {
				continue
			}

			split, needLen, err = msg.ServerHandshakeDecode(ctx, recordType, recordLen, totalBuf[newOffset:newOffset+recordLen])
			if err != nil {
				return nil, err
			}

			if msg.FatalAlert || msg.ServerHelloDone {
				break
			}

			newOffset += recordLen

			//如果确认是已经分包
			if split {
				originNeedLen := needLen
				needLen -= recordLen
				for needLen > 0 {
					if total < newOffset+recordHeaderLen {
						goto _Continue
					}

					//总长度比偏移+记录值+5要长
					_, fragmentLen, err := msg.ServerHandshakeHeaderDecode(ctx, totalBuf[newOffset:newOffset+recordHeaderLen])
					if err != nil {
						return nil, err
					}

					if total < newOffset+recordHeaderLen+fragmentLen {
						goto _Continue
					}

					//剔除碎片Record头部
					totalBuf = append(totalBuf[:newOffset], totalBuf[newOffset+recordHeaderLen:]...)
					newOffset += fragmentLen
					total -= recordHeaderLen
					needLen -= fragmentLen
				} // end for

				if total < newOffset {
					goto _Continue
				}

				//如果满足
				_, _, err = msg.ServerHandshakeDecode(ctx, recordType, originNeedLen, totalBuf[newOffset-originNeedLen:newOffset])
				if err != nil {
					return nil, err
				}
				if msg.FatalAlert || msg.ServerHelloDone {
					break
				}
			} // end split

			current = newOffset

			if total > newOffset {
				goto Again
			}
		}

	_Continue:

		if count == 0 && n == 0 && (err == io.EOF || strings.Contains(err.Error(), "read: connection reset by peer")) {
			return nil, myerr.ErrServerClose
		}

		if err != nil {
			return nil, err
		}

		if n == 0 {
			return nil, myerr.ErrReadZeroData
		}
		count++

	}

	return msg, nil
}

//生成ClientHello中的32位随机数（4byte时间+28byte随机）
func fillRandom(buf []byte) []byte {
	uTime := time.Now().Unix()
	r := rand.New(rand.NewSource(uTime))
	// 4 bytes Unix Time
	binary.BigEndian.PutUint32(buf[:4], uint32(uTime))

	//填充随机数
	for i := 4; i < 32; i++ {
		buf[i] = byte(r.Intn(0xFF))
	}

	return buf
}
