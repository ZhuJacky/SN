package ssl

import (
	"fmt"
	"testing"
	"context"
	"ssldemo/core/myconn"
)


func BenchmarkGetServerHandshakeMsg(b *testing.B) {
	opt := &ClientHelloOptions{
		ServerName: "child-care-preschool.brighthorizons.com",
		MaxVersion: TLSv12,
	}

	param := &myconn.CheckParams{
		Ip:   "121.199.38.96",
		Port: "443",
	}

	for i:=0;i<b.N;i++ {
		msg, err := GetServerHandshakeMsg(context.Background(), param, opt, []CipherID{TLS_RSA_WITH_3DES_EDE_CBC_SHA})
		if err ==nil {
			fmt.Printf("%v",msg)
		}
	}
}
