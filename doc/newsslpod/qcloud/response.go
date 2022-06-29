// Package qcloud provides ...
package qcloud

import "strings"

// Response the qcloud the response
type Response struct {
	// 版本号
	Version int `json:"version"`
	// 被调方模块名
	ComponentName string `json:"componentName"`
	// 时间戳
	Timestamp int64 `json:"timestamp"`
	// 请求ID
	EventId int64 `json:"eventId"`
	// 请求错误码
	ReturnValue int `json:"returnValue"`
	// 请求是否成功
	ReturnCode int `json:"returnCode"`
	// 请求执行信息
	ReturnMessage string `json:"returnMessage"`
	// 请求API
	Interface string `json:"interface,omitempty"`
	// 结果数据
	Data interface{} `json:"data"`
}

// GetBool bool value
func (r *Response) GetBool(field string) (b bool) {
	b, _ = r.getValue(field).(bool)
	return
}

// GetString string value
func (r *Response) GetString(field string) (s string) {
	s, _ = r.getValue(field).(string)
	return
}

// GetStringSlice []string slice
func (r *Response) GetStringSlice(field string) (ss []string) {
	ss, _ = r.getValue(field).([]string)
	return
}

// GetInt int value
func (r *Response) GetInt(field string) (i int) {
	f, _ := r.getValue(field).(float64)
	return int(f)
}

// GetIntSlice []int value
func (r *Response) GetIntSlice(field string) (is []int) {
	fs, _ := r.getValue(field).([]float64)
	for _, v := range fs {
		is = append(is, int(v))
	}
	return
}

// GetInt64 int64 value
func (r *Response) GetInt64(field string) (i64 int64) {
	f, _ := r.getValue(field).(float64)
	return int64(f)
}

// GetFloat64 float64 value
func (r *Response) GetFloat64(field string) (f float64) {
	f, _ = r.getValue(field).(float64)
	return
}

// GetStringMap map[string]interface{}
func (r *Response) GetStringMap(field string) (m map[string]interface{}) {
	m, _ = r.getValue(field).(map[string]interface{})
	return
}

// GetStringMapString map[string]string
func (r *Response) GetStringMapString(field string) (m map[string]string) {
	m, _ = r.getValue(field).(map[string]string)
	return
}

func (r *Response) getValue(field string) interface{} {
	keys := strings.Split(field, ".")

	m := r.Data.(map[string]interface{})
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
