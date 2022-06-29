package myconn

import (
	"context"
	"errors"
	"net"
)

func GetConn(ctx context.Context, checkParams *CheckParams) (conn net.Conn, err error) {
	switch checkParams.ServerType {
	case Web:
		return NewWithContext(ctx, "tcp", checkParams.Ip+":"+checkParams.Port)
	case SMTP:
		return HandshakeWithMail(ctx, checkParams)
	case IMAP:
		return HandshakeWithMail(ctx, checkParams)
	case POP3:
		return HandshakeWithMail(ctx, checkParams)
	default:
		return nil, errors.New("未支持的检测种类(web,smtp,imap,pop3)")

	}
}
