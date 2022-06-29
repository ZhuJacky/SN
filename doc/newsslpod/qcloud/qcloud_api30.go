// Package qcloud provides ...
package qcloud

import "encoding/json"

// 前端请求API框架数据格式
type RequestFront struct {
	Action      string                 `json:"action"`
	Data        map[string]interface{} `json:"data"`
	ServiceType string                 `json:"serviceType"`
}

// API框架请求后端数据格式
type RequestBack map[string]interface{}

func (r RequestBack) GetInt(field string) int {
	f, _ := r[field].(float64)
	return int(f)
}

func (r RequestBack) GetString(field string) string {
	s, _ := r[field].(string)
	return s
}

func (r RequestBack) GetBool(field string) bool {
	b, _ := r[field].(bool)
	return b
}

func (r RequestBack) SetKeys(key string, val interface{}) {
	r[key] = val
}

func (r RequestBack) GetValue(key string) interface{} {
	return r[key]
}

func (r RequestBack) GetRawData() []byte {
	return r["rawData"].([]byte)
}

// 后端返回数据格式
type ResponseBack struct {
	Response interface{}
}

type ResponseError struct {
	RequestId string
	Error     Error
}

type Error struct {
	Code    string
	Message string
}

func ErrorResponse(errCode, reqId string) []byte {
	resp := &ResponseBack{
		Response: ResponseError{
			RequestId: reqId,
			Error: Error{
				Code:    errCode,
				Message: ErrDesc[errCode],
			},
		},
	}
	data, _ := json.Marshal(resp)
	return data
}

// 返回给前端结构
type ResponseFront struct {
	Code int             `json:"code"`
	Data json.RawMessage `json:"data"`
}
