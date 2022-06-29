package myconn

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDetectionServerType(t *testing.T) {
	assert := assert.New(t)
	tests := []struct {
		prot       string
		serverType ServerType
	}{
		{"110", POP3},
		{"995", POP3},
		{"25", SMTP},
		{"465", SMTP},
		{"587", SMTP},
		{"994", SMTP},
		{"143", IMAP},
		{"993", IMAP},
		{"8080", Web},
		{"8081", Web},
	}

	for _, test := range tests {
		result := DetectionServerType(test.prot)
		assert.Equal(test.serverType, result)
	}
}

func TestServerTypeToString(t *testing.T) {
	assert := assert.New(t)

	tests := []struct {
		serverType ServerType
		result     string
	}{
		{Web, "Web"},
		{SMTP, "SMTP"},
		{IMAP, "IMAP"},
		{POP3, "POP"},
	}

	for _, test := range tests {
		result := ServerTypeToString(test.serverType)
		assert.Equal(test.result, result)

	}
}
