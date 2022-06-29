package ssl

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRemoveAssignEllipticCurve(t *testing.T) {
	assert := assert.New(t)
	test := []struct {
		curves []SupportedGroup
		assign SupportedGroup
		result []SupportedGroup
	}{
		{[]SupportedGroup{SECT163K1,
			SECT163R1,
			SECT163R2,
			SECT193R1,
			SECT193R2,
			SECT233K1,
			SECT233R1,
			SECT239K1,
			SECT283K1,
			SECT283R1,
			SECT409K1,
			SECT409R1,
			SECT571K1,
			SECT571R1,
			SECP160K1,
			SECP160R1,
			SECP160R2,
			SECP192K1,
			SECP192R1,
			SECP224K1,
			SECP224R1,
			SECP256K1,
			SECP256R1,
			SECP384R1,
			SECP521R1,
			BRAINPOOLP256R1,
			BRAINPOOLP384R1,
			BRAINPOOLP512R1,
			ECDH_X25519,
			ECDH_X448,
			FFDHE2048,
			FFDHE3072,
			FFDHE4096,
			FFDHE6144,
			FFDHE8192,
		}, SECT163K1, []SupportedGroup{
			SECT163R1,
			SECT163R2,
			SECT193R1,
			SECT193R2,
			SECT233K1,
			SECT233R1,
			SECT239K1,
			SECT283K1,
			SECT283R1,
			SECT409K1,
			SECT409R1,
			SECT571K1,
			SECT571R1,
			SECP160K1,
			SECP160R1,
			SECP160R2,
			SECP192K1,
			SECP192R1,
			SECP224K1,
			SECP224R1,
			SECP256K1,
			SECP256R1,
			SECP384R1,
			SECP521R1,
			BRAINPOOLP256R1,
			BRAINPOOLP384R1,
			BRAINPOOLP512R1,
			ECDH_X25519,
			ECDH_X448,
			FFDHE2048,
			FFDHE3072,
			FFDHE4096,
			FFDHE6144,
			FFDHE8192,
		}},
		{[]SupportedGroup{
			SECT163K1,
			SECT163R1,
			SECT163R2,
			SECT193R1,
			SECT193R2,
			SECT233K1,
			SECT233R1,
			SECT239K1,
			SECT283K1,
			SECT283R1,
			SECT409K1,
			SECT409R1,
			SECT571K1,
			SECT571R1,
			SECP160K1,
			SECP160R1,
			SECP160R2,
			SECP192K1,
			SECP192R1,
			SECP224K1,
			SECP224R1,
			SECP256K1,
			SECP256R1,
			SECP384R1,
			SECP521R1,
			BRAINPOOLP256R1,
			BRAINPOOLP384R1,
			BRAINPOOLP512R1,
			ECDH_X25519,
			ECDH_X448,
			FFDHE2048,
			FFDHE3072,
			FFDHE4096,
			FFDHE6144,
			FFDHE8192,
		}, FFDHE8192, []SupportedGroup{
			SECT163K1,
			SECT163R1,
			SECT163R2,
			SECT193R1,
			SECT193R2,
			SECT233K1,
			SECT233R1,
			SECT239K1,
			SECT283K1,
			SECT283R1,
			SECT409K1,
			SECT409R1,
			SECT571K1,
			SECT571R1,
			SECP160K1,
			SECP160R1,
			SECP160R2,
			SECP192K1,
			SECP192R1,
			SECP224K1,
			SECP224R1,
			SECP256K1,
			SECP256R1,
			SECP384R1,
			SECP521R1,
			BRAINPOOLP256R1,
			BRAINPOOLP384R1,
			BRAINPOOLP512R1,
			ECDH_X25519,
			ECDH_X448,
			FFDHE2048,
			FFDHE3072,
			FFDHE4096,
			FFDHE6144,
		}},
	}

	for _, t := range test {
		result, err := RemoveAssignEllipticCurve(t.curves, t.assign)
		assert.Nil(err)
		assert.Equal(t.result, result)
	}
}

func TestGetAllEllipticCurve(t *testing.T) {
	assert := assert.New(t)
	data := []byte{
		0x00, 0x01, //SECT163K1
		0x00, 0x02, //SECT163R1
		0x00, 0x03, //SECT163R2
		0x00, 0x04, //SECT193R1
		0x00, 0x05, //SECT193R2
		0x00, 0x06, //SECT233K1
		0x00, 0x07, //SECT233R1
		0x00, 0x08, //SECT239K1
		0x00, 0x09, //SECT283K1
		0x00, 0x0a, //SECT283R1
		0x00, 0x0b, //SECT409K1
		0x00, 0x0c, //SECT409R1
		0x00, 0x0d, //SECT571K1
		0x00, 0x0e, //SECT571R1
		0x00, 0x0f, //SECP160K1
		0x00, 0x10, //SECP160R1
		0x00, 0x11, //SECP160R2
		0x00, 0x12, //SECP192K1
		0x00, 0x13, //SECP192R1
		0x00, 0x14, //SECP224K1
		0x00, 0x15, //SECP224R1
		0x00, 0x16, //SECP256K1
		0x00, 0x17, //SECP256R1
		0x00, 0x18, //SECP384R1
		0x00, 0x19, //SECP521R1
		0x00, 0x1a, //BRAINPOOLP256R1
		0x00, 0x1b, //BRAINPOOLP384R1
		0x00, 0x1c, //BRAINPOOLP512R1
		0x00, 0x1d, //ECDH_X25519
		0x00, 0x1e, //ECDH_X448
		0x01, 0x00, //FFDHE2048
		0x01, 0x01, //FFDHE3072
		0x01, 0x02, //FFDHE4096
		0x01, 0x03, //FFDHE6144
		0x01, 0x04, //FFDHE8192
	}
	curve := GetTLSSupportGroup()

	assert.Equal(data, curve)
}

func TestSupportedGroup_ToRawBytes(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		data   SupportedGroup
		result []byte
	}{
		{SECT163K1, []byte{0x00, 0x01}},
		{FFDHE8192, []byte{0x01, 0x04}},
	}

	for _, t := range tests {
		result := t.data.ToRawBytes()
		assert.Equal(result, t.result, t.data)
	}
}

func TestGetSupportGroupFromRaw(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		data   []byte
		result SupportedGroup
	}{
		{[]byte{0x00, 0x01}, SECT163K1},
		{[]byte{0x01, 0x04}, FFDHE8192},
	}

	for _, t := range tests {
		result, _ := GetSupportGroupFromRaw(t.data)
		assert.Equal(result, t.result)
	}
}

func TestSupportedGroup_GetInfo(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		data   SupportedGroup
		exist  bool
		result *SupportedGroupInfo
	}{
		{SECT163K1, true, &SupportedGroupInfo{
			Name:     "sect163k1",
			Data:     SECT163K1,
			RSAEqual: 1024,
		}},
		{SECT193R1, true, &SupportedGroupInfo{
			Name:     "sect193r1",
			Data:     SECT193R1,
			RSAEqual: 1536,
		}},
		{0x0a0b, false, nil},
	}

	for _, t := range tests {
		info, exist := t.data.GetInfo()
		assert.Equal(exist, t.exist)
		if exist {
			assert.Equal(info.Name, t.result.Name)
		}
	}
}
