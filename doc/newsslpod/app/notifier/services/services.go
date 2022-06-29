// Package services provides ...
package services

import (
	"net/http"
	"time"

	"mysslee_qcloud/app/notifier/db"
	"mysslee_qcloud/message"
	"mysslee_qcloud/model"
	"mysslee_qcloud/redis"

	"github.com/gin-gonic/gin"
)

// BasicFilter the basic filter
func BasicFilter(c *gin.Context) {

}

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

func HandleNoticeManual(c *gin.Context) {
	msg := &model.NoticeMsg{
		Uin:      "524602999",
		Type:     15,
		Msg:      `{"Nickname":"nickname","Domain":"test.example.com","Port":"443","Time":"2006-01-02 15:04:05","TrustStatus":"不正常"}`,
		Language: "zh",
		// NoticeAlready: model.NoticeNone,
		CreatedAt: time.Now(),
	}
	err := db.AddNoticeMsg(msg)
	if err != nil {
		c.String(http.StatusOK, err.Error())
	} else {
		c.String(http.StatusOK, "ok")
	}
}
