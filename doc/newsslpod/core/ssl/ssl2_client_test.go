package ssl

import (
	"context"
	"testing"

	"mysslee_qcloud/core/myconn"

	"github.com/stretchr/testify/assert"
)

// func ExampleGenerateClientHello() {
// 	opt := &SSL2ClientHelloOpt{}
// 	opt.Ciphers = GetOnlySSL2Ciphers()
// 	result, err := opt.GenerateClientHello(false, 0)
// 	if err != nil {
// 		log.Panicf("err:%v", err)
// 	}
// 	fmt.Printf("0x%x\n", result[0])     //type
// 	fmt.Printf("0x%x\n", result[1])     //Length
// 	fmt.Printf("0x%x\n", result[2])     //Client Hello
// 	fmt.Printf("0x%x\n", result[3:5])   //Version
// 	fmt.Printf("0x%x\n", result[5:7])   //Cipher Spec Length
// 	fmt.Printf("0x%x\n", result[7:9])   //Session ID Length
// 	fmt.Printf("0x%x\n", result[9:11])  // Challenge length
// 	fmt.Printf("0x%x\n", result[11:14]) //套件一
// 	fmt.Printf("0x%x\n", result[14:17]) //套件二
// 	fmt.Printf("0x%x\n", result[17:20]) //套件三
// 	fmt.Printf("0x%x\n", result[20:23]) //套件四
// 	fmt.Printf("0x%x\n", result[23:26]) //套件五
// 	fmt.Printf("0x%x\n", result[26:29]) //套件六
// 	fmt.Printf("0x%x\n", result[29:32]) ///套件七
// 	fmt.Printf("0x%x\n", result[32:])   //Challenge
//
// 	//Output:
// 	//0x80
// 	//0x2e
// 	//0x1
// 	//0x0002
// 	//0x0015
// 	//0x0000
// 	//0x0010
// 	//0x010080
// 	//0x020080
// 	//0x030080
// 	//0x040080
// 	//0x050080
// 	//0x060040
// 	//0x0700c0
// 	//0x00010203040506070001020304050607
//
// }

func TestSSL2ClientHelloOpt_GetServerHandshakeMsg(t *testing.T) {
	assert := assert.New(t)
	conn, err := myconn.NewWithContext(context.Background(), "tcp", "180.167.86.2:10001")
	assert.Nil(err)
	if err == nil {

		opt := SSL2ClientHelloOpt{}
		msg, err := opt.GetServerHandshakeMsg(context.Background(), conn, false, 0, GetOnlySSL2Ciphers())
		assert.Nil(err)
		assert.Equal([]byte{0x01, 0x00, 0x80, 0x03, 0x00, 0x80, 0x05, 0x00, 0x80, 0x07, 0x00, 0xC0}, msg.SSL2ServerHandshakeMsg.Cipher)
	}
}
