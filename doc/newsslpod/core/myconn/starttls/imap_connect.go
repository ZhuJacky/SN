//impa 邮件连接
package starttls

import (
	"errors"
	"net"
	"strings"

	"mysslee_qcloud/core/myerr"
)

func DoImapStarttls(conn net.Conn) (err error) {

	buf := make([]byte, 1024)
	var totalBuf []byte
	var total int
	var ok bool

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			totalBuf = append(totalBuf, buf[:n]...)
			total += n
			if total > 4 {
				result := string(totalBuf[:4])
				if strings.Contains(strings.ToLower(result), "ok") {
					ok = true
					break
				}
				break
			}
		}
		if err != nil {
			return err
		}

		if n == 0 {
			return myerr.ErrReadZeroData
		}
	}

	if !ok {
		return errors.New("无法建立连接")
	}

	_, err = conn.Write([]byte("a001 starttls\r\n")) //写命令
	if err != nil {
		return err
	}

	totalBuf = make([]byte, 0)
	total = 0
	ok = false

	for {
		n, err := conn.Read(buf)
		if n > 0 {
			totalBuf = append(totalBuf, buf[:n]...)
			total += n
			if total > 7 {
				if strings.Contains(strings.ToLower(string(totalBuf[:7])), "ok") {
					ok = true
				}
				break
			}

		}
		if err != nil {
			return err
		}
		if n == 0 {
			return myerr.ErrReadZeroData
		}
	}

	if !ok {
		return errors.New("不支持StartTLS")
	}
	return nil
}
