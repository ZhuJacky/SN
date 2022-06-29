// Package myconn TODO
package myconn

import (
	"context"
	"crypto/tls"
	"errors"
	"mysslee_qcloud/dns"
	"net"
	"net/http"
	"strings"
	"time"

	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
)

// 一些变量
var (
	SSConf                        SSConfig
	cipher                        *ss.Cipher
	TransportSkipVerify           *http.Transport
	TransportSkipVerifyAndNoCache *http.Transport
)

// SSConfig TODO
type SSConfig struct {
	Open    bool
	Servers []struct {
		Addr         string
		LocationCode string
	}
	ServerPort int
	Password   string
	Method     string
}

// GetByLocationCode TODO
func (s *SSConfig) GetByLocationCode(lc string) (addr string) {
	for i := 0; i < len(s.Servers); i++ {
		if s.Servers[i].LocationCode == lc {
			return s.Servers[i].Addr
		}
	}
	return ""
}

func init() {
	http.DefaultTransport = &http.Transport{
		Proxy:                 nil,           // 禁用HTTP代理
		DialContext:           myDialContext, // 替换为自己的Dial函数
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
	}

	TransportSkipVerify = &http.Transport{
		Proxy:                 nil,           // 禁用HTTP代理
		DialContext:           myDialContext, // 替换为自己的Dial函数
		MaxIdleConns:          100,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   10 * time.Second,
		ExpectContinueTimeout: 1 * time.Second,
		TLSClientConfig:       &tls.Config{InsecureSkipVerify: true},
	}

	TransportSkipVerifyAndNoCache = &http.Transport{
		Proxy:             nil,           // 禁用HTTP代理
		DialContext:       myDialContext, // 替换为自己的Dial函数
		TLSClientConfig:   &tls.Config{InsecureSkipVerify: true},
		DisableKeepAlives: true, // 禁用连接缓存
	}
}

var myDialer = &net.Dialer{
	Timeout:   10 * time.Second,
	KeepAlive: 30 * time.Second,
	DualStack: true,
}

func myDialContext(ctx context.Context, network, address string) (net.Conn, error) {
	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	ip := net.ParseIP(host)
	// 私有 IP 或者 本地主机名(docker容器间依赖)
	if (ip != nil && dns.IsPrivateV4(ip)) || !strings.Contains(host, ".") {
		return myDialer.DialContext(ctx, network, address)
	}

	addrs := []string{host}
	// 非 IP，解析域名
	if ip == nil {
		addrs, err = dns.LookupHost(ctx, host)
		if err != nil {
			return nil, err
		}
		if len(addrs) == 0 {
			return nil, errors.New("no resolve to address " + address)
		}
	}
	// judge internal ip addr
	if dns.IsPrivateIP(addrs[0]) {
		return nil, errors.New("unsupported internet ip address")
	}

	return myDialer.DialContext(ctx, network, addrs[0]+":"+port)
}

// Init TODO
func Init(c SSConfig) {
	SSConf = c

	var err error
	cipher, err = ss.NewCipher(SSConf.Method, SSConf.Password)
	if err != nil {
		panic(err)
	}
}

// New TODO
func New(network, addr string) (conn net.Conn, err error) {
	return NewWithContext(context.Background(), network, addr)
}

// NewWithContext TODO
func NewWithContext(ctx context.Context, network, addr string) (conn net.Conn, err error) {
	if ctx == nil {
		return nil, errors.New("nil context")
	}

	// 正常拨号
	dialer := &net.Dialer{
		Timeout: 10 * time.Second, // 连接超时
	}
	conn, err = dialer.DialContext(ctx, network, addr)
	if err != nil {
		return nil, err
	}

	if ctx != context.Background() {
		// 设置主任务超时时间
		if time, ok := ctx.Deadline(); ok {
			conn.SetDeadline(time)
		}
	}

	return
}
