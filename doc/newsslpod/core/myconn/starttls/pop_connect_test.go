//POP3 邮件连接
package starttls

import (
	"fmt"
	"net"
	"testing"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestDialPop3ServerWithStarttls(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		domain string
		ip     string
		port   string
		result bool
	}{
		{"pop.163.com", "220.181.12.110", "110", true},
	}

	for _, test := range tests {
		addr := fmt.Sprintf("%v:%v", test.ip, test.port)
		conn, err := net.Dial("tcp", addr)
		if err != nil {
			log.Panic("获取连接错误")
		}
		err = DoPop3Starttls(conn)
		if test.result {
			assert.Nil(err)
		} else {
			assert.NotNil(err)
		}

	}

}
