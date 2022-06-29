//smtp邮件连接
package starttls

import (
	"errors"
	"net"
	"net/textproto"
	"strings"
)

type client struct {
	Text      *textproto.Conn
	Conn      net.Conn
	ext       map[string]string
	localName string
}

func DoSmtpStarttls(conn net.Conn) (err error) {

	c, err := newClient(conn)
	if err != nil {
		return err
	}
	err = c.ehlo()
	if err != nil {
		return err
	}

	if _, ok := c.ext["STARTTLS"]; !ok {
		return errors.New("该邮件服务器不支持STARTTLS")
	}

	_, err = c.starttls()
	if err != nil {
		return err
	}

	return nil
}

func newClient(conn net.Conn) (*client, error) {
	text := textproto.NewConn(conn)
	_, _, err := text.ReadResponse(220) //拨号成功，邮件服务器响应220
	if err != nil {
		text.Close()
		return nil, err
	}
	c := &client{Text: text, Conn: conn, localName: "MySSL.com"}
	return c, nil
}

//复制 系统包smtp包中的数据
func (c *client) cmd(expectCode int, format string, args ...interface{}) (code int, msg string, err error) {
	id, err := c.Text.Cmd(format, args...)
	if err != nil {
		return 0, "", err
	}
	c.Text.StartResponse(id)
	defer c.Text.EndResponse(id)
	code, msg, err = c.Text.ReadResponse(expectCode)
	return code, msg, err
}

//发送ehlo命令
func (c *client) ehlo() error {
	_, msg, err := c.cmd(250, "EHLO %s", c.localName)
	if err != nil {
		return err
	}
	ext := make(map[string]string)
	extList := strings.Split(msg, "\n")
	if len(extList) > 1 {
		extList = extList[1:]
		for _, line := range extList {
			args := strings.SplitN(line, " ", 2)
			if len(args) > 1 {
				ext[args[0]] = args[1]
			} else {
				ext[args[0]] = ""
			}
		}
	}
	c.ext = ext
	return nil
}

//开始tls握手部分
func (c *client) starttls() (conn net.Conn, err error) {
	_, _, err = c.cmd(220, "STARTTLS") //开始TSL握手
	if err != nil {
		return nil, err
	}
	return c.Conn, nil
}
