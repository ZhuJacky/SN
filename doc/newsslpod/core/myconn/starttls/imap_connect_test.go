package starttls

import (
	"fmt"
	"net"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

//看wireshark流程
func TestDialImapServerWithStarttls(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		domain string
		ip     string
		port   string
		result bool
	}{
		{"imap.163.com", "220.181.12.100", "143", true},
		{"imap.qq.com", "219.133.60.187", "142", false},
	}

	for _, test := range tests {
		addr := fmt.Sprintf("%v:%v", test.ip, test.port)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Panic("获取连接错误")
		}
		err = DoImapStarttls(conn)
		if test.result {
			assert.Nil(err)
		} else {
			assert.NotNil(err)
		}
	}
}
