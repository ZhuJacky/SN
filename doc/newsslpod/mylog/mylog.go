// Package mylog provides ...
package mylog

import (
	"bytes"
	"fmt"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var logger *logrus.Logger

func Init(path string) {
	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logger = &logrus.Logger{
		Out:          file,
		Formatter:    &MyFormatter{},
		ReportCaller: true,
		Level:        logrus.InfoLevel,
	}
}

func Info(costTime time.Duration, name string, request *http.Request, reqBody, respBody []byte, args ...interface{}) {
	logger.WithFields(logrus.Fields{
		"costTime":  costTime,
		"interface": name,
		"request":   request,
		"reqBody":   reqBody,
		"respBody":  respBody,
	}).Info(args...)
}

type MyFormatter struct{}

func (f *MyFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	// time
	t := entry.Time.Format("15:04:05")
	f.WriteString(b, t)
	// level
	l, err := entry.Level.MarshalText()
	if err != nil {
		return nil, err
	}
	f.Write(b, bytes.ToUpper(l))
	// file
	slugs := strings.Split(entry.Caller.File, "/")
	f.WriteString(b, slugs[len(slugs)-1]+":"+fmt.Sprint(entry.Caller.Line))
	// send response
	req := entry.Data["request"].(*http.Request)
	f.WriteString(b, fmt.Sprintf("send response: %s %s costtime=%v",
		req.Method, entry.Data["interface"].(string),
		entry.Data["costTime"].(time.Duration)))
	// client ip
	h, _, err := net.SplitHostPort(req.RemoteAddr)
	if err != nil {
		return nil, err
	}
	f.WriteString(b, fmt.Sprintf("client=%s", h))
	// request body
	f.Write(b, entry.Data["reqBody"].([]byte))
	// response body
	f.Write(b, entry.Data["respBody"].([]byte))
	f.WriteString(b, entry.Message)

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func (f *MyFormatter) Write(b *bytes.Buffer, data []byte) {
	b.Write(data)
	b.WriteByte(' ')
}

func (f *MyFormatter) WriteString(b *bytes.Buffer, str string) {
	b.WriteString(str)
	b.WriteByte(' ')
}

// 00:00:01 DEBUG MY_Output:90 1 5cec0980c3900: send response: POST /backend/domain costtime=374 client=9.66.19.233 request body: [{"version":1,"componentName":"dnspod-web","eventId":15
// 58972800726.8,"seqId":"95d0d9b1a870bfa131edc8da3d704ce9","interface":{"interfaceName":"wss.domain.pureCheckDomainReg","para":{"domain":"blb.com.cn"}}}] response body: [{"version":"1.
// 0","componentName":"wss","timestamp":1558972801,"eventId":1558972800726.8,"returnValue":0,"returnCode":0,"returnMessage":"success","data":{"reg_status":1,"domain":"blb.com.cn","domai
// n_name":"blb","tld":".com.cn"}}]

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Stop timer
		end := time.Now()
		latency := end.Sub(start)

		if raw != "" {
			path = path + "?" + raw
		}
		var req, resp []byte
		if body := c.Keys["reqbody"]; body != nil {
			req = body.([]byte)
		}
		if body := c.Keys["respbody"]; body != nil {
			resp = body.([]byte)
		}
		logrus.Infof("%v %s reqeust body=%s response body=%s", latency, path, string(req), string(resp))
	}
}
