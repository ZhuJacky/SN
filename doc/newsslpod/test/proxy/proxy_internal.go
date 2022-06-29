// Package proxy provides ...
package proxy

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"mysslee_qcloud/qcloud"
	"mysslee_qcloud/test/mc"
	"mysslee_qcloud/test/sslpod"
)

func ProxyInternal(c *gin.Context) {
	fmt.Println(">>>internal proxy: /api/internal/call")

	data, err := c.GetRawData()
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	// unmarshal
	req := &qcloud.Request{}
	err = json.Unmarshal(data, req)
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	interfaceName := req.Interface.InterfaceName
	service := strings.Split(interfaceName, ".")[1]

	var resp *qcloud.Response
	switch service {
	case "message":
		resp, err = mc.HandleRequest(req)
	case "sslpod":
		resp, err = sslpod.HandleRequestInternal(req)
	}
	if err != nil {
		c.String(http.StatusInternalServerError, err.Error())
		return
	}
	c.JSON(http.StatusOK, resp)
}
