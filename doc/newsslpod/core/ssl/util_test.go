package ssl

import (
	"fmt"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestStringToCiphers(t *testing.T) {
	assert := assert.New(t)
	ciphers, unSupportCiphers := StringToCiphers("00ffc02cc02bc024c023c00ac009c008c030c02fc028c027c014c013c012009d009c003d003c0035002f000ac007c01100050004")
	for _, c := range ciphers {
		fmt.Println(c)
	}

	assert.Equal(26, len(ciphers), "获取支持的套件数量不正确")

	for i, c := range unSupportCiphers {
		log.Warnf("unsupport index %v,data:%v", i, c)
	}
	assert.Equal(0, len(unSupportCiphers), "获取不支持的套件的数量不正确")
}

func TestCiphersFromHex(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		data    string
		ciphers []CipherID
	}{
		{"00010003", []CipherID{TLS_RSA_WITH_NULL_MD5, TLS_RSA_EXPORT_WITH_RC4_40_MD5}},
	}

	for _, test := range tests {
		result := CiphersFromHex(test.data)
		assert.Equal(test.ciphers, result)
	}
}
