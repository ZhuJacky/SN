package myconn

import (
	"context"
	"errors"
	"fmt"
	"net"
	"time"

	"mysslee_qcloud/core/myconn/starttls"
)

func HandshakeWithMail(ctx context.Context, checkParams *CheckParams) (conn net.Conn, err error) {
	addr := fmt.Sprintf("%v:%v", checkParams.Ip, checkParams.Port)
	conn, err = NewWithContext(ctx, "tcp", addr)
	if err != nil {
		return nil, err
	}

	if checkParams.MailDirect {
		return conn, nil
	}

	var connReady bool
	defer func() {
		if !connReady {
			conn.Close()
		}
	}()

	switch checkParams.ServerType {
	case SMTP:
		err = starttls.DoSmtpStarttls(conn)
	case IMAP:
		err = starttls.DoImapStarttls(conn)
	case POP3:
		err = starttls.DoPop3Starttls(conn)
	default:
		return nil, errors.New("不支持的检测类型，默认的邮件端口")
	}

	connReady = err == nil
	return conn, err
}

//测试是否能够直接使用TLS进行握手
func CheckMailDirectSSL(ctx context.Context, checkParams *CheckParams) (mailDirect bool, err error) {

	addr := fmt.Sprintf("%v:%v", checkParams.Ip, checkParams.Port)
	conn, err := NewWithContext(ctx, "tcp", addr)
	if err != nil {
		return false, err
	}

	conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	buf := make([]byte, 1024)

	n, err := conn.Read(buf)

	if n > 0 {
		return false, nil
	}

	if err != nil {
		switch err.(type) {
		case *net.OpError:
			e, ok := err.(*net.OpError)
			if !ok {
				return false, err
			}

			if e.Timeout() {
				return true, nil
			}
		}
	}

	return false, errors.New("无法判断邮件服务器是否启用starttls")
}
