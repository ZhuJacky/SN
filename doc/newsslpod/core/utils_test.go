package core

import (
	"context"
	"testing"
)

func TestParseHost(t *testing.T) {
	var hosts = []string{
		"127.0.0.1",
		"127.0.0.1:80",
		"baidu.com",
		"baidu.com:80",
		"[::1]:80",
		"2001:db8::68",
		"[2001:db8::68:12]:80",

		"2404:6800:4008:803::200e",
		"[2404:6800:4008:803::200e]:8080",
		"115.29.145.169",
		"115.29.145.169:80",
	}

	for _, v := range hosts {
		h, d, p, ip, dt, err := ParseHost(context.Background(), v)
		if err != nil {
			t.Error(v, err)
			continue
		}

		t.Log(h, d, p, ip, dt)
	}
}
