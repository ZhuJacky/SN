// Package message provides ...
package message

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type Message struct {
	Code  int         `json:"code"`  //状态码
	Error string      `json:"error"` //错误信息
	Data  interface{} `json:"data"`  //信息
}

// json 消息
func JSON(c *gin.Context, msg *Message) {
	//如果code不是成功，并且错误信息为空
	if msg.Error == "" && msg.Code != Success {
		msg.Error = CodeDesc[msg.Code]
	}

	//如果code是成功，并且正确信息为空
	if msg.Code == Success && msg.Data == nil {
		msg.Data = CodeDesc[msg.Code]
	}

	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "GET, POST, PUT,OPTIONS")
	c.JSON(http.StatusOK, msg)
}
