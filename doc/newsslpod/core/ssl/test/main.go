package main

import (
	"context"
	//"mysslee_qcloud/core/myconn"
	//log "github.com/sirupsen/logrus"
	ssl "mysslee_qcloud/core/ssl"
	"fmt"
	"net"
)

func main() {
	opt := &ssl.ClientHelloOptions{
		ServerName: "www.myssl.com",
		MaxVersion: ssl.TLSv12,
	}
	conn, err := net.Dial("tcp", "54.223.206.183:443")
	if err != nil {
		fmt.Printf("errconn")
		//log.Panicf("获取连接错误:%v", err)
	}

	msg, err := ssl.GetServerHandShakeMsgWithConn(context.Background(), conn, opt, ssl.GoDefaultCiphers)
	fmt.Printf("abbc %v",msg)
	if err == nil {
		//hello, _ := msg.Hello()
		fmt.Printf("abbc %v",msg)
		//log.Infof("version:0x%x \n", hello.Version)
		//log.Infof("cipher: %v", hello.Cipher)
	}
}

func add(a int,b int) int {
	return a+b
}