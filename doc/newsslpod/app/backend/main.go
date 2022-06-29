// Package main provides ...
package main

import (
	"mysslee_qcloud/app/backend/db"
	"mysslee_qcloud/app/backend/router"
	"mysslee_qcloud/app/backend/timer"
	"mysslee_qcloud/config"
	"mysslee_qcloud/kafka"
	"mysslee_qcloud/polaris"
	"mysslee_qcloud/redis"
	"mysslee_qcloud/utils"
	"net/http"
	_ "net/http/pprof"

	"github.com/sirupsen/logrus"
)

// @title MySSL EE Backend API
// @version 1.0
// @description 为 qcloud 提供 openapi 功能.

// @contact.name henry.chen
// @contact.email henry.chen@trustasia.com

// @host localhost:20000
// @BasePath /api
func init() {
	config.InitBackend()

	redis.Init(redis.BackendApp)
	db.Init()
	kafka.InitProducer()
	polaris.RegisterBackend()
	timer.Init()

	// 先将日志注释掉，使用tke标准输入输出流记录日志到cls
	// file, err := os.OpenFile(config.Conf.Backend.LogPath,
	// 	os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// logrus.SetOutput(file)
}

func main() {
	defer func() {
		if err := recover(); err != nil {
			logrus.Panic(utils.CallerStack())
		}
	}()

	if config.Conf.RunMode != config.MODE_PRE_PRODUCTION &&
		config.Conf.RunMode != config.MODE_PRODUCTION {
		go func() {
			http.ListenAndServe(":6060", nil)
		}()
	}

	router.Run()
}
