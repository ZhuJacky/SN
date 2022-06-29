package starttls

import (
	"fmt"
	"net"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDialMailServerContext(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		domain string
		ip     string
		port   string
	}{
		{"smtp.qq.com", "14.18.245.164", "587"},
	}

	for _, test := range tests {
		addr := fmt.Sprintf("%v:%v", test.ip, test.port)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Panic("获取连接错误")
		}

		err = DoSmtpStarttls(conn)
		assert.Nil(err, test.domain)
	}
}
