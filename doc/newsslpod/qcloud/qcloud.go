// Package qcloud provides ...
package qcloud

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"mysslee_qcloud/config"
	"mysslee_qcloud/model"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// qcloudHandler handle the qcloud request
type (
	qcloudHandler func(RequestBack) (*ResponseBack, string)
	apiHandler    func(*Request) (interface{}, string)
)

var (
	exHandlers = make(map[string]qcloudHandler)
	inHandlers = make(map[string]apiHandler)
	lock       sync.Mutex
)

// Register qcloud handler
func Register(apiName string, api interface{}) {
	lock.Lock()
	defer lock.Unlock()

	switch handler := api.(type) {
	case func(RequestBack) (*ResponseBack, string):
		if _, ok := exHandlers[apiName]; ok {
			panic("duplicate register handler: " + apiName)
		}
		exHandlers[apiName] = handler
	case func(*Request) (interface{}, string):
		if _, ok := inHandlers[apiName]; ok {
			panic("duplicate register handler: " + apiName)
		}
		inHandlers[apiName] = handler
	default:
		panic("unrecognize handler type")
	}
}

// HandleExternalCall handle the api request
func HandleExternalCall(c *gin.Context, initAcct func(uin, ip string) (*model.Account, error)) {
	reqData, err := c.GetRawData()
	if err != nil {
		logrus.Error("HandleExternalCall.GetRawData: ", err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}

	// unmarshal
	req := RequestBack{}
	err = json.Unmarshal(reqData, &req)
	if err != nil {
		logrus.Error("HandleExternalCall.Unmarshal: ", err)
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	req.SetKeys("rawData", reqData)

	var (
		errCode string
		resp    *ResponseBack
	)
	defer func() {
		var respData []byte
		if errCode != "" {
			respData = ErrorResponse(errCode, req.GetString("RequestId"))
		} else {
			respData, _ = json.Marshal(resp)
		}
		c.Data(http.StatusOK, "application/json", respData)

		// log
		c.Set("reqbody", reqData)
		c.Set("respbody", respData)
	}()

	// initial account
	account, err := initAcct(req.GetString("Uin"), c.ClientIP())
	if err != nil {
		logrus.Error("HandleExternalCall.initAcct: ", err)
		errCode = ErrInternalError
		return
	}
	req.SetKeys("account", account)
	// get handler
	handler, ok := exHandlers[req.GetString("Action")]
	if !ok {
		errCode = ErrInvalidAction
		return
	}
	// exec handler
	resp, errCode = handler(req)
}

// HandleInternalCall handle the api request
func HandleInternalCall(c *gin.Context) {
	reqData, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// unmarshal
	req := &Request{}
	err = json.Unmarshal(reqData, req)
	if err != nil {
		c.String(http.StatusBadRequest, err.Error())
		return
	}
	// response
	var errCode string
	resp := &Response{
		Version:       1,
		EventId:       req.EventId,
		ComponentName: "QC_" + strings.ToUpper(config.Conf.QCloud.Module),
		ReturnCode:    0,
	}
	defer func() {
		if errCode != "" {
			resp.ReturnValue = 1
			resp.ReturnMessage = errCode
		} else {
			resp.ReturnMessage = "ok"
		}
		respData, _ := json.Marshal(resp)
		c.Status(http.StatusOK)
		c.Writer.Write(respData)

		// log
		c.Set("reqbody", reqData)
		c.Set("respbody", respData)
	}()
	// check
	interfaceName := req.Interface.InterfaceName
	service := strings.Split(interfaceName, ".")
	if len(service) != 3 {
		errCode = ErrInvalidParameterValue
		return
	}
	if service[1] != config.Conf.QCloud.Module {
		errCode = ErrUnsupportedOperation
		return
	}
	// get handler
	handler, ok := inHandlers[service[2]]
	if !ok {
		errCode = ErrInvalidAction
		return
	}
	// exec handler
	resp.Data, errCode = handler(req)
	resp.Timestamp = time.Now().Unix()
}

// GetAccountInfoFromQCloud send request for get account info.
func GetAccountInfoFromQCloud(uin string) (*Response, error) {
	req := &Request{
		Version:       1,
		Timestamp:     time.Now().Unix(),
		ComponentName: "MC",
		EventId:       rand.Int63(),
		Interface: Interface{
			InterfaceName: "qcloud.Quser.getNickname",
			Para: map[string]interface{}{
				"uin": uin,
			},
		},
	}
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}
	// request
	response, err := http.Post(config.Conf.QCloud.AccountGateway, "application/json", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()
	if response.StatusCode != 200 {
		return nil, errors.New(response.Status)
	}
	// read data
	data, err = ioutil.ReadAll(response.Body)
	if err != nil {
		return nil, err
	}
	resp := &Response{}
	// Unmarshal
	err = json.Unmarshal(data, resp)
	if err != nil {
		return nil, err
	}
	if resp.ReturnValue != 0 {
		return nil, errors.New(resp.ReturnMessage)
	}
	return resp, nil
}
