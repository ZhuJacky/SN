// Package qcloud provides ...
package qcloud

import (
	"strings"
)

// Interface the Interface struct
type Interface struct {
	InterfaceName string                 `json:"interfaceName"`
	Para          map[string]interface{} `json:"para"`
}

// Request the qcloud the request
type Request struct {
	// 版本号
	Version int `json:"version"`
	// 调用方
	Caller string `json:"caller,omitempty"`
	// 调用方模块名
	ComponentName string `json:"componentName"`
	// 序列ID
	SeqId string `json:"seqId,omitempty"`
	// SpanId
	SpanId string `json:"spanId,omitempty"`
	// 请求ID
	EventId int64 `json:"eventId"`
	// 时间戳
	Timestamp int64 `json:"timestamp,omitempty"`
	// 接口详情
	Interface Interface `json:"interface"`
	// 自定义数据
	Keys map[string]interface{} `json:"-"`
}

// SetKeys set key value
func (r *Request) SetKeys(key string, value interface{}) {
	if r.Keys == nil {
		r.Keys = make(map[string]interface{})
	}
	r.Keys[key] = value
}

// GetBool bool value
func (r *Request) GetBool(field string) (b bool) {
	b, _ = r.getValue(field).(bool)
	return
}

// GetString string value
func (r *Request) GetString(field string) (s string) {
	s, _ = r.getValue(field).(string)
	return
}

// GetStringSlice []string slice
func (r *Request) GetStringSlice(field string) (ss []string) {
	is, _ := r.getValue(field).([]interface{})
	ss = make([]string, len(is))
	for i, v := range is {
		ss[i], _ = v.(string)
	}
	return
}

// GetInt int value
func (r *Request) GetInt(field string) (i int) {
	f, _ := r.getValue(field).(float64)
	return int(f)
}

// GetIntSlice []int value
func (r *Request) GetIntSlice(field string) (is []int) {
	// NOTE it's invalid
	fs, _ := r.getValue(field).([]float64)
	for _, v := range fs {
		is = append(is, int(v))
	}
	return
}

// GetInt64 int64 value
func (r *Request) GetInt64(field string) (i64 int64) {
	f, _ := r.getValue(field).(float64)
	return int64(f)
}

// GetFloat64 float64 value
func (r *Request) GetFloat64(field string) (f float64) {
	f, _ = r.getValue(field).(float64)
	return
}

// GetStringMap map[string]interface{}
func (r *Request) GetStringMap(field string) (m map[string]interface{}) {
	m, _ = r.getValue(field).(map[string]interface{})
	return
}

// GetStringMapString map[string]string
func (r *Request) GetStringMapString(field string) (m map[string]string) {
	m, _ = r.getValue(field).(map[string]string)
	return
}

func (r *Request) getValue(field string) interface{} {
	keys := strings.Split(field, ".")

	m := r.Interface.Para
	for i, k := range keys {
		if val, ok := m[k]; ok {
			if i == len(keys)-1 {
				return val
			}
			m, ok = val.(map[string]interface{})
			if !ok {
				return nil
			}
		}
	}
	return nil
}
