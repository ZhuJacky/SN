// Package router provides ...
package router

import (
	"fmt"

	"mysslee_qcloud/app/notifier/services"
	"mysslee_qcloud/config"
	"mysslee_qcloud/prom"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var router *gin.Engine

func init() {
	if config.Conf.RunMode == config.MODE_PRE_PRODUCTION ||
		config.Conf.RunMode == config.MODE_PRODUCTION {
		gin.SetMode(gin.ReleaseMode)

		logrus.SetLevel(logrus.InfoLevel)
	}

	router = gin.Default()
	router.GET("/notice", services.HandleNoticeManual)
	// router
	api := router.Group("/api", services.BasicFilter)
	{
		api.GET("/app/infos", services.HandleAppInfos)
	}
	// prometheus
	router.GET("/metrics", prom.HandlePrometheus)
}

// Run start gin
func Run() {
	router.Run(":" + fmt.Sprint(config.Conf.Notifier.Listen))
}
