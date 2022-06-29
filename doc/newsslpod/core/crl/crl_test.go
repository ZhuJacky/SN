package crl

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckCRL(t *testing.T) {
	assert := assert.New(t)
	data, err := ioutil.ReadFile("test/ca.pem")
	assert.Nil(err, "读取CA证书错误")
	block, _ := pem.Decode(data)
	cert, err := x509.ParseCertificate(block.Bytes)
	assert.Nil(err, "解析证书错误")
	os.Chdir("../../webapp")
	revoke, err := CheckCRL(context.Background(), cert)
	assert.Nil(err, "检测CRL错误")
	assert.Equal(false, revoke, "证书CRL吊销情况错误")
}
