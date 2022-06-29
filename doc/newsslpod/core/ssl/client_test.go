package ssl

import (
	"context"

	"testing"

	"mysslee_qcloud/core/myconn"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)



func TestGetServerHandshakeMsg(t *testing.T) {
	assert := assert.New(t)
	opt := &ClientHelloOptions{
		ServerName: "child-care-preschool.brighthorizons.com",
		MaxVersion: TLSv12,
	}

	param := &myconn.CheckParams{
		Ip:   "208.99.183.5",
		Port: "465",
	}

	msg, err := GetServerHandshakeMsg(context.Background(), param, opt, []CipherID{TLS_RSA_WITH_3DES_EDE_CBC_SHA})
	assert.Nil(err)
	if err == nil {
		hello, _ := msg.Hello()
		log.Infof("version:0x%x \n", hello.Version)
		log.Infof("cipher: %v", hello.Cipher)
	}
}
