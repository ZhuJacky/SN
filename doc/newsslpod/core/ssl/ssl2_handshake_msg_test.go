package ssl

import (
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var ssl2CipherData = []byte{0x01, 0x00, 0x80, 0x07, 0x00, 0xC0}

func TestGetSSL2CiphersInfo(t *testing.T) {
	assert := assert.New(t)

	infos, err := GetSSL2CiphersInfo(ssl2CipherData)
	assert.Nil(err)

	assert.Equal(2, len(infos))
	for _, info := range infos {
		log.Infof(info.Name)
	}
}
