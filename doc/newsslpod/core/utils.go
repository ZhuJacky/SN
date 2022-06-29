package core

import (
	"context"
	"errors"
	"net"
	"strings"
	"time"

	"mysslee_qcloud/dns"
	"mysslee_qcloud/utils"

	"golang.org/x/net/idna"
)

// 切分 s
// host host:port [host]:port [ipv6-host%zone]:port
// eg. 2001:db8::68
// eg. [::1]:80
// eg. baidu.com
// eg. baidu.com:80
func ParseHost(ctx context.Context, s string) (domain, punycodeDomain, port string, ips []string, dnsTime int64, err error) {
	// 将域名改成小写
	s = strings.ToLower(s)

	// 切分域名
	domain, port, err = net.SplitHostPort(s)
	if err != nil {
		if strings.Contains(err.Error(), "missing port") {
			err = nil
			domain = s
			port = "443"
		} else {
			err = errors.New("请输入正确的域名")
			return
		}
	}

	// 检查端口
	if !utils.CheckPort(port) { //验证端口的合法性
		err = errors.New("请输入正确的端口")
		return
	}

	// 如果是 IP
	if ip := net.ParseIP(domain); ip != nil {
		// 判断是否是IPv4
		if strings.Contains(domain, ":") {
			err = errors.New("请输入IPv4地址")
			return
		}

		ips = []string{domain}
		punycodeDomain = domain
	} else if punycodeDomain, err = idna.ToASCII(domain); err == nil {
		// 如果是域名
		if utils.ValidateDomain2(punycodeDomain) {
			err = errors.New("不正确的域名")
			return
		}
		now := time.Now()
		ips, err = dns.LookupHost(ctx, punycodeDomain)
		if err != nil {
			return
		}

		if len(ips) == 0 {
			err = errors.New("没有找到IP")
			return
		}

		dnsTime = time.Now().Sub(now).Nanoseconds() / 10e6
	} else {
		err = errors.New("不正确的域名或IP")
		return
	}

	// 是否是IP和私有IP
	if dns.IsPrivateIP(ips[0]) {
		err = errors.New("请输入域名／公网IP")
		return
	}

	return
}
