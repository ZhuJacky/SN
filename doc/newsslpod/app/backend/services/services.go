// Package services provides ...
package services

import (
	"encoding/gob"
	"mysslee_qcloud/config"
	"mysslee_qcloud/message"
	"mysslee_qcloud/model"
	"mysslee_qcloud/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

// Internal filter
var InternalFilter = gin.BasicAuth(gin.Accounts(config.Conf.Backend.BasicAuth))

// HandleAppInfos redis app infos
// @Summary app 相关信息
// @Description app信息
// @Tags 内部调用
// @Accept mpfd
// @Produce json
// @Success 200 {object} message.Message
// @Router /app/infos [get]
func HandleAppInfos(c *gin.Context) {
	msg := &message.Message{}
	defer message.JSON(c, msg)

	backendIPs, err := redis.ScanApp(redis.BackendApp)
	if err != nil {
		msg.Code = message.ErrSystem
		msg.Error = err.Error()
		return
	}
	checkerIPs, err := redis.ScanApp(redis.CheckerApp)
	if err != nil {
		msg.Code = message.ErrSystem
		msg.Error = err.Error()
		return
	}
	notifierIPs, err := redis.ScanApp(redis.NotifierApp)
	if err != nil {
		msg.Code = message.ErrSystem
		msg.Error = err.Error()
		return
	}

	msg.Code = message.Success
	msg.Data = gin.H{
		"backend":  backendIPs,
		"checker":  checkerIPs,
		"notifier": notifierIPs,
	}
}

// HandleTaskCallback 接受kafka任务检测完之后的回调
func HandleTaskCallback(c *gin.Context) {
	msg := &model.CallbackToBackend{}

	err := gob.NewDecoder(c.Request.Body).Decode(&msg)
	if err != nil {
		logrus.Error("HandleTaskCallback.NewDecoder: ", err)
		return
	}
	return
}
