// Package ct 该包主要内容为解析证书或TLS中的SCT扩展
package ct

import (
	"encoding/asn1"
	"errors"

	"mysslee_qcloud/utils"

	ct "github.com/google/certificate-transparency-go"
	ctls "github.com/google/certificate-transparency-go/tls"
)

// TODO： 将原来零散的一些解析迁移过来

var (
	// X509SCTExtensionID SCT扩展ID
	X509SCTExtensionID = []int{1, 3, 6, 1, 4, 1, 11129, 2, 4, 2}

	// TLSSCTExtensionHeader TLS SCT 扩展头
	TLSSCTExtensionHeader = []byte{0x00, 0x12}
)

var (
	// ErrExtNil 扩展为空的错误
	ErrExtNil = errors.New("except x509 extension nil")
	// ErrNotSCTExt 非SCT扩展错误
	ErrNotSCTExt = errors.New("seems not a sct extension")
)

// SerializedSCT sct扩展
//	  opaque SerializedSCT<1..2^16-1>;
type SerializedSCT struct {
	Val []byte `tls:"minlen:1,maxlen:65535"`
}

// SignedCertificateTimestampList SCT列表
//	  struct {
//         SerializedSCT sct_list <1..2^16-1>;
//    } SignedCertificateTimestampList;
type SignedCertificateTimestampList struct {
	SctList []SerializedSCT `tls:"minlen:1,maxlen:65535"`
}

// ParseX509SCTExtension 解析证书内SCT扩展
// Extension Id: 2b 06 01 04 01 d6 79 02 04 02  (1.3.6.1.4.1.11129.2.4.2)
// 			  04 81 f1
// Length:    00 ef
// SCT1:  	  00 76 00 29 3c 51 96 54 c8 39 65 ba aa 50 fc 58
// 			  07 d4 b7 6f bf 58 7a 29 72 dc a4 c3 0c f4 e5 45
// 			  47 f4 78 00 00 01 64 99 9e 44 a6 00 00 04 03 00
// 			  47 30 45 02 20 67 e8 f5 34 64 be 4e 1b ed f7 05
// 			  e4 01 3a 6a d1 e5 9c cf 9f 75 a1 04 c7 27 e3 23
// 			  fa a2 88 57 35 02 21 00 c4 46 e4 ea 14 83 73 c4
// 			  37 ce df e8 33 e2 4e 10 4b 54 ea 43 3d 51 4c 45
// 			  2c 1f 32 44 8f 6e 3a 26
// SCT2:...
//
func ParseX509SCTExtension(sctExt []byte) (scts []ct.SignedCertificateTimestamp, err error) {
	if len(sctExt) <= 3 {
		return nil, ErrExtNil
	}

	index := 0

	// 头部 0x04 表示OCTET string
	octstringLen := 1
	if int(sctExt[index]) != asn1.TagOctetString {
		return nil, ErrNotSCTExt
	}
	index += octstringLen

	// 取扩展长度
	// sctLen < 0x80
	// sctLen < 0xFF  0x81xx
	// sctLen < 0xFFFF 0x82xxxx
	// sctLen < 0xFFFFFF 0x83xxxxxx
	headerLen := 1
	var sctLengthLen int
	switch sctExt[index] {
	case 0x81:
		sctLengthLen = 1
	case 0x82:
		sctLengthLen = 2
	case 0x83:
		sctLengthLen = 3
	default:
		headerLen = 0
		sctLengthLen = 1
	}
	index += headerLen

	if len(sctExt) < octstringLen+headerLen+sctLengthLen {
		return nil, ErrNotSCTExt
	}

	sctLen := utils.BytesToInt(sctExt[index : index+sctLengthLen])

	// 判断扩展的长度是否正确
	if sctLen != len(sctExt)-octstringLen-headerLen-sctLengthLen {
		return nil, ErrNotSCTExt
	}

	index += sctLengthLen

	return unmarshalSCTs(sctExt[index:])
}

// ParseTLSSCTExtension 解析从TLS扩展中来的SCT
// header   00 12
// extLen	00 f0
// sctLen	00 ee
// sct1		00 75 00 db 74 af ee cb 29 ec b1 fe ca 3e 71 6d
//			2c e5 b9 aa bb 36 f7 84 71 83 c7 5d 9d 4f 37 b6
// 			1f bf 64 00 00 01 64 99 bf ba 00 30 99 00 00 04
// 			03 00 46 30 44 02 20 0b 74 40 d4 42 00 40 03 43
// 			5a 01 ff 86 41 46 c8 28 fa 14 96 78 18 bf 00 50
// 			4d a7 8d 3c 92 6c 15 b8 29 f2 ab 02 20 79 e7 34
// 			00 60 6d 3d 18 90 4a 02 f2 b9 23 22 f8 67 80 77
// 			f3 55 00 70 77 46 fd 09 75 1a 54 67 d1 c1 73 ce
// 			5f
// sct2: ..
func ParseTLSSCTExtension(sctExt []byte) (scts []ct.SignedCertificateTimestamp, err error) {
	if len(sctExt) <= 4 {
		return nil, ErrExtNil
	}

	index := 0

	// 解析时 已将头部去了
	// sctHeaderLen := 2
	// if !bytes.Equal(sctExt[index:sctHeaderLen], TLSSCTExtensionHeader) {
	// 	return nil, ErrNotSCTExt
	// }
	//
	// index += sctHeaderLen
	sctExtLengthLen := 2
	if utils.BytesToInt(sctExt[index:index+sctExtLengthLen]) != len(sctExt)-sctExtLengthLen {
		return nil, ErrNotSCTExt
	}

	index += sctExtLengthLen

	return unmarshalSCTs(sctExt[index:])
}

func unmarshalSCTs(raw []byte) (scts []ct.SignedCertificateTimestamp, err error) {
	sctList := SignedCertificateTimestampList{}
	_, err = ctls.Unmarshal(raw, &sctList)
	if err != nil {
		return
	}

	for _, serializedSCT := range sctList.SctList {
		if len(serializedSCT.Val) <= 2 {
			err = ErrExtNil
			return
		}
		sct := ct.SignedCertificateTimestamp{}
		_, err = ctls.Unmarshal(serializedSCT.Val, &sct)
		if err != nil {
			return
		}
		scts = append(scts, sct)
	}
	return
}
