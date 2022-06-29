// Package sslpod provides ...
package sslpod

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"

	"mysslee_qcloud/qcloud"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var MYSSLEE_BACKEND_ADDR = "http://127.0.0.1:20000"

func init() {
	addr := os.Getenv("MYSSLEE_BACKEND_ADDR")
	if addr != "" {
		MYSSLEE_BACKEND_ADDR = addr
	}
}

func HandleRequestExternal(c *gin.Context, req qcloud.RequestBack) ([]byte, string) {
	data, _ := json.Marshal(req)
	resp, err := http.Post(MYSSLEE_BACKEND_ADDR+"/external/call",
		"application/json", bytes.NewReader(data))
	if err != nil {
		logrus.Error("HandleRequestExternal.Post: ", err)
		return nil, qcloud.ErrInternalError
	}
	defer resp.Body.Close()

	data, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, qcloud.ErrInternalError
	}

	return data, ""
}

func HandleRequestInternal(req *qcloud.Request) (resp *qcloud.Response, err error) {
	data, err := json.Marshal(req)
	if err != nil {
		return resp, nil
	}
	result, err := http.Post(MYSSLEE_BACKEND_ADDR+"/internal/call",
		"application/json", bytes.NewReader(data))
	if err != nil {
		return
	}
	defer result.Body.Close()

	data, err = ioutil.ReadAll(result.Body)
	if err != nil {
		return
	}
	if result.StatusCode != 200 {
		err = errors.New("Status code: " + result.Status)
		return
	}
	resp = new(qcloud.Response)
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}
	return
}
