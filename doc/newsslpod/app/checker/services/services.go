// Package services provides ...
package services

import (
	"encoding/gob"
	"encoding/json"

	"mysslee_qcloud/app/checker/check"
	"mysslee_qcloud/config"
	"mysslee_qcloud/kafka"
	"mysslee_qcloud/message"
	"mysslee_qcloud/model"
	"mysslee_qcloud/redis"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func init() {
	gob.Register([]model.DomainResult{})
	kafka.InitConsumer()
}

func BasicFilter() gin.HandlerFunc {
	return gin.BasicAuth(gin.Accounts(config.Conf.Checker.BasicAuth))
}

// Internal filter
var InternalFilter = gin.BasicAuth(gin.Accounts{
	"qcloud": "bab658e3c1176a664ccbdd74a4f4b9a5",
})

// HandleAppInfos etcd app infos
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

// HandleTaskCheck
// @Summary 获取检测任务
// @Description 检测任务
// @Tags 内部调用
// @Accept mpfd
// @Produce json
// @Success 200 {object} message.Message
// @Router /task [post]
func HandleTaskCheck(c *gin.Context) {
	msg := &message.Message{}
	defer message.JSON(c, msg)

	var result []model.DomainResult

	err := gob.NewDecoder(c.Request.Body).Decode(&result)
	if err != nil {
		logrus.Error("HandleTaskCheck.NewDecoder: ", err)
		msg.Code = message.ErrIncorrectParams
		return
	}
	for _, v := range result {
		check.DomainChecker.DoFast(v)
	}
	msg.Code = message.Success
}

// InitKafkaCheck 启动kafka检测
func InitKafkaCheck() {
	go HanleKafkaChecks()
}

// HanleKafkaChecks 进行kafka检测
func HanleKafkaChecks() {
	for {
		msg, err := kafka.ConsumeMessage()
		if err != nil {
			logrus.Error("HandleTaskCheck.ConsumeMessage: ", err)
			continue
		}
		var re []model.KafkaDomainInfo
		err = json.Unmarshal(msg.Value, &re)
		if err != nil {
			logrus.Error("HandleTaskCheck.NewDecoder: ", err)
			continue
		}
		for _, v := range re {
			check.DomainChecker.DoKafka(v)
		}
	}
}

// DoKafkaAndRecover 做kafka检测，异常恢复
func DoKafkaAndRecover(v model.KafkaDomainInfo) {
	defer func() {
		if err := recover(); err != nil {
			logrus.Panic(err)
		}
	}()
	check.DomainChecker.DoKafka(v)
}
