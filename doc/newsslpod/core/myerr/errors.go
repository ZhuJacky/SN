package myerr

import (
	"io"
	"net"
	"strings"
)

// 一些通用错误
var (
	ErrGetCertNoChain = &MError{OutErr: "证书链为空"}
	ErrReadZeroData   = &MError{OutErr: "获取握手错误，读取到非法的长度0！"}
	ErrServerClose    = &MError{OutErr: "服务器断开连接"}
)

// MError core custom error
type MError struct {
	NetErr error    // 原始的网络错误
	OutErr string   // 该Err会输出到用户层
	OthErr []string // 描述性错误, 不输出
}

// Error 可输出的Error
func (e *MError) Error() string {
	if e != nil && e.OutErr != "" {
		return e.OutErr
	}

	if e != nil && e.OthErr != nil {
		return strings.Join(e.OthErr, ",")
	}

	if e != nil && e.NetErr != nil {
		return e.NetErr.Error()
	}

	return ""
}

// GetNetErrorKeywords 关键网络错误
func GetNetErrorKeywords(err error) string {
	if e, ok := err.(*MError); ok && e != nil && e.NetErr != nil {
		// 一般网络错误形式如 connect xxx port xxx failed: Connection refused
		// 按照：分割，取最后一个
		es := strings.Split(e.NetErr.Error(), ":")
		return es[len(es)-1]
	}
	return ""
}

// GetNetError 输出网络错误
func GetNetError(err error) error {
	if e, ok := err.(*MError); ok {
		return e.NetErr
	}
	return nil
}

// GetErrors 输出所有错误
func GetErrors(err error) string {
	if e, ok := err.(*MError); ok && e != nil {
		var out string
		if e.OutErr != "" {
			out = out + ": " + e.OutErr
		}

		if e.OthErr != nil {
			out = out + ": " + strings.Join(e.OthErr, ",")
		}

		if e.NetErr != nil {
			out = out + ": " + e.NetErr.Error()
		}
		return out
	}

	if err != nil {
		return err.Error()
	}

	return ""
}

// ToNetError doc
func ToNetError(err error) *MError {
	if e, ok := err.(*MError); ok {

		// 不覆盖上层err
		if e.NetErr == nil {
			e.NetErr = err
		}
		return e
	}

	return &MError{NetErr: err}
}

// ToOutError doc
// 外层需要添加额外的展示信息，同时不丢掉里面能显示问题的信息
// 将里面的信息降级到otherr 内
func ToOutError(err error, outMsg string) *MError {
	if e, ok := err.(*MError); ok {

		if e.OutErr == "" {
			e.OutErr = outMsg
			return e
		}

		if e.OthErr == nil {
			e.OthErr = []string{e.OutErr}
		} else {
			e.OthErr = append(e.OthErr, e.OutErr)
		}

		e.OutErr = outMsg
		return e
	}

	if err == nil {
		return ErrorOut(outMsg)
	}

	return &MError{OthErr: []string{err.Error()}, OutErr: outMsg}
}

// ErrorOut doc
func ErrorOut(errStr string) *MError {
	if errStr == "" {
		return nil
	}

	return &MError{OutErr: errStr}
}

// TransformError 客户端握手部分 错误格式话输出
func TransformError(err error) string {

	if err == io.EOF {
		return "服务器断开连接"
	}

	switch e := err.(type) {
	case *net.OpError: //网络的错误

		if e.Timeout() {
			return "连接超时"
		}

		if e.Temporary() {
			return "服务器断开连接"
		}

		switch e.Err.(type) {
		case *net.DNSError:
			return "DNS错误"
		}

		if strings.Contains(e.Err.Error(), "connection refused") {
			return "连接被拒绝"
		}

	case *MError:
		if e == ErrServerClose {
			return "服务器断开连接"
		}
	default:
		return "握手错误"
	}
	return err.Error()
}
