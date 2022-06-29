// Package main provides ...
package main

import (
	"net/http"
	_ "net/http/pprof"
	"os"

	"mysslee_qcloud/app/notifier/db"
	"mysslee_qcloud/app/notifier/router"
	"mysslee_qcloud/app/notifier/timer"
	"mysslee_qcloud/config"
	"mysslee_qcloud/redis"
	"mysslee_qcloud/utils"

	"github.com/sirupsen/logrus"
)

// @title MySSL EE Notifier API
// @version 1.0
// @description 为 qcloud 提供 openapi 功能.

// @contact.name henry.chen
// @contact.email henry.chen@trustasia.com

// @host localhost:20020
// @BasePath /api
func init() {
	config.InitNotifier()

	redis.Init(redis.NotifierApp)
	db.Init()

	file, err := os.OpenFile(config.Conf.Notifier.LogPath,
		os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	logrus.SetOutput(file)
}

// start
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
	go timer.Start()
	router.Run()
}
