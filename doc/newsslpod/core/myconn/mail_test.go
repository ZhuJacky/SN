package myconn

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckMailDirectSSL(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		domain string
		ip     string
		port   string
	}{
		{"pop3.sina.com", "113.108.216.75", "110"},
	}

	for _, t := range tests {
		direct, err := CheckMailDirectSSL(context.Background(), &CheckParams{Domain: t.domain, Ip: t.ip, Port: t.port, ServerType: POP3})
		assert.Nil(err)
		assert.Equal(direct, false)
	}

}
