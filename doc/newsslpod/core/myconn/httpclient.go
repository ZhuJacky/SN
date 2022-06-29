package myconn

import (
	"context"
	"crypto/tls"
	"errors"
	"io"
	"net/http"
	"time"
)

const UserAgent = `Mozilla/5.0 (Macintosh; Intel Mac OS X 10_11_5) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/51.0.2704.103 Safari/537.36`

var transport = &http.Transport{
	Proxy:                 nil,                                   // 禁用HTTP代理
	DialContext:           myDialContext,                         // 替换为自己的Dial函数
	TLSClientConfig:       &tls.Config{InsecureSkipVerify: true}, // 关闭TLS验证，供检测使用
	MaxIdleConns:          0,                                     // 关闭连接复用
	IdleConnTimeout:       90 * time.Second,
	TLSHandshakeTimeout:   10 * time.Second,
	ExpectContinueTimeout: 1 * time.Second,
}

// 支持ss策略的HTTP请求，用于检测程序，不验证证书
// 通过IP连接时，指定host可以获取到正确内容
func HttpRequest(ctx context.Context, method, urlStr, host string, maxRedirect int, body io.Reader, header http.Header) (resp *http.Response, err error) {
	req, err := http.NewRequest(method, urlStr, body)
	if err != nil {
		return
	}
	if header != nil {
		req.Header = header
	}
	req.Header.Set("User-Agent", UserAgent) //进行客户端伪装
	if host != "" {
		// req.URL.Host = host // 重定向referer会用到，但是会导致覆盖urlStr指定的IP，权衡不能用此方法！！
		// req.Header.Set("Host", host)
		req.Host = host
	}

	if ctx != nil {
		req = req.WithContext(ctx)
	}

	client := &http.Client{
		Transport: transport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// fmt.Println(req)
			if len(via) > maxRedirect {
				return errors.New("终止重定向.")
			}
			return nil
		},
	}

	return client.Do(req)
}
