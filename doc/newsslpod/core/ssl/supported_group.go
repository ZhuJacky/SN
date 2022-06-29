package ssl

import "errors"

type SupportedGroup uint16

const (
	SECT163K1                       SupportedGroup = 0x0001
	SECT163R1                       SupportedGroup = 0x0002
	SECT163R2                       SupportedGroup = 0x0003
	SECT193R1                       SupportedGroup = 0x0004
	SECT193R2                       SupportedGroup = 0x0005
	SECT233K1                       SupportedGroup = 0x0006
	SECT233R1                       SupportedGroup = 0x0007
	SECT239K1                       SupportedGroup = 0x0008
	SECT283K1                       SupportedGroup = 0x0009
	SECT283R1                       SupportedGroup = 0x000a
	SECT409K1                       SupportedGroup = 0x000b
	SECT409R1                       SupportedGroup = 0x000c
	SECT571K1                       SupportedGroup = 0x000d
	SECT571R1                       SupportedGroup = 0x000e
	SECP160K1                       SupportedGroup = 0x000f
	SECP160R1                       SupportedGroup = 0x0010
	SECP160R2                       SupportedGroup = 0x0011
	SECP192K1                       SupportedGroup = 0x0012
	SECP192R1                       SupportedGroup = 0x0013
	SECP224K1                       SupportedGroup = 0x0014
	SECP224R1                       SupportedGroup = 0x0015
	SECP256K1                       SupportedGroup = 0x0016
	SECP256R1                       SupportedGroup = 0x0017
	SECP384R1                       SupportedGroup = 0x0018
	SECP521R1                       SupportedGroup = 0x0019
	BRAINPOOLP256R1                 SupportedGroup = 0x001a
	BRAINPOOLP384R1                 SupportedGroup = 0x001b
	BRAINPOOLP512R1                 SupportedGroup = 0x001c
	ECDH_X25519                     SupportedGroup = 0x001d
	ECDH_X448                       SupportedGroup = 0x001e
	FFDHE2048                       SupportedGroup = 0x0100
	FFDHE3072                       SupportedGroup = 0x0101
	FFDHE4096                       SupportedGroup = 0x0102
	FFDHE6144                       SupportedGroup = 0x0103
	FFDHE8192                       SupportedGroup = 0x0104
	Reserved                        SupportedGroup = 0x0a0a // TODO: 应该是随机的
	ARBITRARY_EXPLICIT_PRIME_CURVES SupportedGroup = 0xff01 //表示支持任意椭圆
	ARBITRARY_EXPLICIT_CHAR2_CURVES SupportedGroup = 0xff02 //表示支持任意特色二曲线

)

func GetSupportGroupFromRaw(data []byte) (group SupportedGroup, err error) {

	if len(data) != 2 {
		return 0, errors.New("无法生成SupportedGroup数据")
	}

	return SupportedGroup(uint16(data[0])<<8) | SupportedGroup(data[1]), nil
}

func (s SupportedGroup) ToRawBytes() []byte {

	data := make([]byte, 2)
	data[0] = byte(s >> 8)
	data[1] = byte(s & 0xff)
	return data
}

func (s SupportedGroup) GetInfo() (info *SupportedGroupInfo, exist bool) {
	info, exist = ellipticCurves[s]
	return
}

type SupportedGroupInfo struct {
	Name     string //椭圆名称
	RSAEqual int    //相当于RSA的强度
	Data     SupportedGroup
}

var ellipticCurves map[SupportedGroup]*SupportedGroupInfo

func init() {
	ellipticCurves = make(map[SupportedGroup]*SupportedGroupInfo, 37)
	ellipticCurves[SECT163K1] = &SupportedGroupInfo{
		Name:     "sect163k1",
		Data:     SECT163K1,
		RSAEqual: 1024,
	}
	ellipticCurves[SECT163R1] = &SupportedGroupInfo{
		Name:     "sect163r1",
		Data:     SECT163R1,
		RSAEqual: 1024,
	}
	ellipticCurves[SECT163R2] = &SupportedGroupInfo{
		Name:     "sect163r2",
		Data:     SECT163R2,
		RSAEqual: 1024,
	}
	ellipticCurves[SECT193R1] = &SupportedGroupInfo{
		Name:     "sect193r1",
		Data:     SECT193R1,
		RSAEqual: 1536,
	}
	ellipticCurves[SECT193R2] = &SupportedGroupInfo{
		Name:     "sect193r2",
		Data:     SECT193R2,
		RSAEqual: 1536,
	}
	ellipticCurves[SECT233K1] = &SupportedGroupInfo{
		Name:     "sect233k1",
		Data:     SECT233K1,
		RSAEqual: 2240,
	}
	ellipticCurves[SECT233R1] = &SupportedGroupInfo{
		Name:     "sect233r1",
		Data:     SECT233R1,
		RSAEqual: 2240,
	}
	ellipticCurves[SECT239K1] = &SupportedGroupInfo{
		Name:     "sect239k1",
		Data:     SECT239K1,
		RSAEqual: 2304,
	}
	ellipticCurves[SECT283K1] = &SupportedGroupInfo{
		Name:     "sect283k1",
		Data:     SECT283K1,
		RSAEqual: 3456,
	}
	ellipticCurves[SECT283R1] = &SupportedGroupInfo{
		Name:     "sect283r1",
		Data:     SECT283R1,
		RSAEqual: 3456,
	}
	ellipticCurves[SECT409K1] = &SupportedGroupInfo{
		Name:     "sect409k1",
		Data:     SECT409K1,
		RSAEqual: 7680,
	}
	ellipticCurves[SECT409R1] = &SupportedGroupInfo{
		Name:     "sect409r1",
		Data:     SECT409R1,
		RSAEqual: 7680,
	}
	ellipticCurves[SECT571K1] = &SupportedGroupInfo{
		Name:     "sect571k1",
		Data:     SECT571K1,
		RSAEqual: 15360,
	}
	ellipticCurves[SECT571R1] = &SupportedGroupInfo{
		Name:     "sect571r1",
		Data:     SECT571R1,
		RSAEqual: 15360,
	}

	ellipticCurves[SECP160K1] = &SupportedGroupInfo{
		Name:     "secp160k1",
		Data:     SECP160K1,
		RSAEqual: 1024,
	}
	ellipticCurves[SECP160R1] = &SupportedGroupInfo{
		Name:     "secp160r1",
		Data:     SECP160R1,
		RSAEqual: 1024,
	}
	ellipticCurves[SECP160R2] = &SupportedGroupInfo{
		Name:     "secp160r2",
		Data:     SECP160R2,
		RSAEqual: 1024,
	}
	ellipticCurves[SECP192K1] = &SupportedGroupInfo{
		Name:     "secp192k1",
		Data:     SECP192K1,
		RSAEqual: 1536,
	}
	ellipticCurves[SECP192R1] = &SupportedGroupInfo{
		Name:     "secp192r1",
		Data:     SECP192R1,
		RSAEqual: 1536,
	}
	ellipticCurves[SECP224K1] = &SupportedGroupInfo{
		Name:     "secp224k1",
		Data:     SECP224K1,
		RSAEqual: 2048,
	}
	ellipticCurves[SECP224R1] = &SupportedGroupInfo{
		Name:     "secp224r1",
		Data:     SECP224R1,
		RSAEqual: 2048,
	}
	ellipticCurves[SECP256K1] = &SupportedGroupInfo{
		Name:     "secp256k1",
		Data:     SECP256K1,
		RSAEqual: 3072,
	}
	ellipticCurves[SECP256R1] = &SupportedGroupInfo{
		Name:     "secp256r1",
		Data:     SECP256R1,
		RSAEqual: 3072,
	}
	ellipticCurves[SECP384R1] = &SupportedGroupInfo{
		Name:     "secp384r1",
		Data:     SECP384R1,
		RSAEqual: 7680,
	}
	ellipticCurves[SECP521R1] = &SupportedGroupInfo{
		Name:     "secp521r1",
		Data:     SECP521R1,
		RSAEqual: 15360,
	}

	ellipticCurves[BRAINPOOLP256R1] = &SupportedGroupInfo{
		Name:     "brainpoolp256r1",
		Data:     BRAINPOOLP256R1,
		RSAEqual: 3072,
	}
	ellipticCurves[BRAINPOOLP384R1] = &SupportedGroupInfo{
		Name:     "brainpoolp384r1",
		Data:     BRAINPOOLP384R1,
		RSAEqual: 7680,
	}
	ellipticCurves[BRAINPOOLP512R1] = &SupportedGroupInfo{
		Name:     "brainpoolp512r1",
		Data:     BRAINPOOLP512R1,
		RSAEqual: 15360,
	}
	ellipticCurves[ECDH_X25519] = &SupportedGroupInfo{
		Name:     "x25519",
		Data:     ECDH_X25519,
		RSAEqual: 3072,
	}
	ellipticCurves[ECDH_X448] = &SupportedGroupInfo{
		Name:     "x448",
		Data:     ECDH_X448,
		RSAEqual: 3072,
	}
	ellipticCurves[FFDHE2048] = &SupportedGroupInfo{
		Name:     "ffdhe2048",
		Data:     FFDHE2048,
		RSAEqual: 2048,
	}
	ellipticCurves[FFDHE3072] = &SupportedGroupInfo{
		Name:     "ffdhe3072",
		Data:     FFDHE3072,
		RSAEqual: 3072,
	}
	ellipticCurves[FFDHE4096] = &SupportedGroupInfo{
		Name:     "ffdhe4096",
		Data:     FFDHE4096,
		RSAEqual: 4096,
	}
	ellipticCurves[FFDHE6144] = &SupportedGroupInfo{
		Name:     "ffdhe6144",
		Data:     FFDHE6144,
		RSAEqual: 6144,
	}
	ellipticCurves[FFDHE8192] = &SupportedGroupInfo{
		Name:     "ffdhe8192",
		Data:     FFDHE8192,
		RSAEqual: 8192,
	}

	//TODO 确定Reserved相对于RSA的强度
	ellipticCurves[Reserved] = &SupportedGroupInfo{
		Name:     "reserved",
		Data:     Reserved,
		RSAEqual: 3072,
	}
}

var tLS13SupportedGroup []SupportedGroup
var tLSSupportedGroup []SupportedGroup

func GetTLS13SupportGroup() []SupportedGroup {
	result := make([]SupportedGroup, len(tLS13SupportedGroup))
	copy(result, tLS13SupportedGroup)
	return result
}

func GetTLSSupportGroup() []SupportedGroup {
	result := make([]SupportedGroup, len(tLSSupportedGroup))
	copy(result, tLSSupportedGroup)
	return result
}

func init() {
	tLSSupportedGroup = getAllSupportedGroup()
}

func getAllSupportedGroup() []SupportedGroup { //根据ssllabs的抓包，最多28个
	var result []SupportedGroup
	result = append(result, SECT163K1)       //1
	result = append(result, SECT163R1)       //2
	result = append(result, SECT163R2)       //3
	result = append(result, SECT193R1)       //4
	result = append(result, SECT193R2)       //5
	result = append(result, SECT233K1)       //6
	result = append(result, SECT233R1)       //7
	result = append(result, SECT239K1)       //8
	result = append(result, SECT283K1)       //9
	result = append(result, SECT283R1)       //10
	result = append(result, SECT409K1)       //11
	result = append(result, SECT409R1)       //12
	result = append(result, SECT571K1)       //13
	result = append(result, SECT571R1)       //14
	result = append(result, SECP160K1)       //15
	result = append(result, SECP160R1)       //16
	result = append(result, SECP160R2)       //17
	result = append(result, SECP192K1)       //18
	result = append(result, SECP192R1)       //19
	result = append(result, SECP224K1)       //20
	result = append(result, SECP224R1)       //21
	result = append(result, SECP256K1)       //22
	result = append(result, SECP256R1)       //23
	result = append(result, SECP384R1)       //24
	result = append(result, SECP521R1)       //25
	result = append(result, BRAINPOOLP256R1) //26
	result = append(result, BRAINPOOLP384R1) //27
	result = append(result, BRAINPOOLP512R1) //28
	result = append(result, ECDH_X25519)     //29
	result = append(result, ECDH_X448)       //30
	result = append(result, FFDHE2048)       //31
	result = append(result, FFDHE3072)       //32
	result = append(result, FFDHE4096)       //33
	result = append(result, FFDHE6144)       //34
	result = append(result, FFDHE8192)       //35
	return result
}

func RemoveAssignEllipticCurve(curves []SupportedGroup, assign SupportedGroup) (newCurves []SupportedGroup, err error) {
	index := -1
	for i := 0; i < len(curves); i++ {
		if curves[i] == assign {
			index = i
			break
		}
	}
	if index == -1 {
		return curves, errors.New("该指定的椭圆不存在")
	}

	curves = append(curves[:index], curves[index+1:]...) //去除指定的椭圆
	return curves, nil
}

func GetEllipticCurvesData(curves []SupportedGroup) []byte {
	result := make([]byte, 0)
	for _, c := range curves {
		result = append(result, c.ToRawBytes()...)
	}
	return result
}
