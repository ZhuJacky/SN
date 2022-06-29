// Package main provides ...
package main

import (
	"net/http"
	_ "net/http/pprof"

	"mysslee_qcloud/app/checker/db"
	"mysslee_qcloud/app/checker/router"
	"mysslee_qcloud/app/checker/services"
	"mysslee_qcloud/config"
	"mysslee_qcloud/polaris"
	"mysslee_qcloud/redis"
	"mysslee_qcloud/utils"

	"github.com/sirupsen/logrus"
)

// @title MySSL EE Checker API
// @version 1.0
// @description 为 qcloud 提供 openapi 功能.

// @contact.name henry.chen
// @contact.email henry.chen@trustasia.com

// @host localhost:20020
// @BasePath /api
func init() {
	config.InitChecker()

	redis.Init(redis.CheckerApp)
	db.Init()
	polaris.RegisterChecker()

	// 先将日志注释掉，使用tke标准输入输出流记录日志到cls
	// file, err := os.OpenFile(config.Conf.Checker.LogPath,
	// 	os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	// if err != nil {
	// 	panic(err)
	// }
	// logrus.SetOutput(file)
	services.InitKafkaCheck()
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
