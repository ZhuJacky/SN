// Package proxy provides ...
package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"mysslee_qcloud/qcloud"
	"mysslee_qcloud/test/billing"
	"mysslee_qcloud/test/sslpod"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// ProxyExternal
func ProxyExternal(c *gin.Context) {
	fmt.Println(">>>enter proxy: /external/api")

	var (
		errCode string
		data    []byte
		resp    qcloud.ResponseFront
		reqId   = fmt.Sprint(time.Now().Unix())
	)
	defer func() {
		if errCode != "" {
			resp.Code = 1
			resp.Data = qcloud.ErrorResponse(errCode, reqId)
		} else {
			resp.Code = 0
			resp.Data = json.RawMessage(data)
		}
		c.JSON(http.StatusOK, resp)
	}()

	data, err := c.GetRawData()
	if err != nil {
		logrus.Error("ProxyExternal.GetRawData: ", err)
		errCode = qcloud.ErrInternalError
		return
	}
	// unmarshal
	req := &qcloud.RequestFront{}
	err = json.Unmarshal(data, req)
	if err != nil {
		logrus.Error("ProxyExternal.Unmarshal: ", string(data))
		errCode = qcloud.ErrMissingParameter
		return
	}
	// NOTE add system parameters
	req.Data["Uin"] = "2223"
	req.Data["AppId"] = "12234"

	reqBack := map[string]interface{}{
		"RequestId": reqId,
		"Action":    req.Action,
		"Uin":       2223,
		"AppId":     12234,
	}
	for k, v := range req.Data {
		reqBack[k] = v
	}
	switch req.ServiceType {
	case "billing":
		data, errCode = billing.HandleQueryPrice(c, reqBack)
	case "sslpod":
		data, errCode = sslpod.HandleRequestExternal(c, reqBack)
	}
}
