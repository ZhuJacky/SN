package utils

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"
)

type ReqInfo struct {
	Id        uint32 // 递增的ID
	Domain    string // 检测的域名
	IP        string
	Port      string
	Who       string    // 请求者IP
	Ext       bool      // 来自浏览器插件
	StartTime time.Time // 请求开始时间
}

type certFinishContextKey struct{}
type reqContextKey struct{}

var gReqInfoId uint32

func (r *ReqInfo) String() string {
	return fmt.Sprintf(`id:%d who:%s domain:%s:%s:%v ext:%v time:%fs`, r.Id, r.Who, r.Domain, r.Port, r.IP, r.Ext, float64(time.Now().Sub(r.StartTime).Nanoseconds())/1000/1000/1000)
}

func MakeReqContext(ctx context.Context, who, domain, port string, ip string, ext bool) context.Context {
	req := &ReqInfo{
		Id:        atomic.AddUint32(&gReqInfoId, 1),
		Who:       who,
		Domain:    domain,
		Port:      port,
		IP:        ip,
		Ext:       ext,
		StartTime: time.Now(),
	}
	return context.WithValue(ctx, reqContextKey{}, req)
}

// IP信息可能是后面解析出来的，需要更新
func UpdateReqContextInfo(ctx context.Context, ip string) bool {
	if ctx != nil {
		if inf := ctx.Value(reqContextKey{}); inf != nil {
			if req, ok := inf.(*ReqInfo); ok {
				req.IP = ip
				return true
			}
		}
	}
	return false
}

func GetReqInfoFromContext(ctx context.Context) string {
	if ctx != nil {
		if inf := ctx.Value(reqContextKey{}); inf != nil {
			if req, ok := inf.(*ReqInfo); ok {
				return req.String()
			}
		}
	}
	return ""
}

//在context中获取证书获取完成
func GetLeafCertFinishContext(ctx context.Context) bool {
	if ctx != nil {
		if info := ctx.Value(certFinishContextKey{}); info != nil {
			if result, ok := info.(bool); ok {
				return result
			}
		}
	}
	return false

}

//设置叶子证书检测完成
func SetLeafCertFinishContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, certFinishContextKey{}, true)
}

func GetReqDuration(ctx context.Context) int {
	if ctx != nil {
		if inf := ctx.Value(reqContextKey{}); inf != nil {
			if req, ok := inf.(*ReqInfo); ok {
				return int(time.Now().Sub(req.StartTime).Seconds())
			}
		}
	}

	return 0
}
