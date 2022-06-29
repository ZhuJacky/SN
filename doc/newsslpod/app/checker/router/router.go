// Package router provides ...
package router

import (
	"fmt"

	"mysslee_qcloud/app/checker/services"
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
	// router
	api := router.Group("/api", services.BasicFilter())
	{
		api.POST("/task", services.HandleTaskCheck)
	}
	internal := router.Group("/api", services.InternalFilter)
	{
		// app info
		internal.GET("/app/infos", services.HandleAppInfos)
	}
	// prometheus
	router.GET("/metrics", prom.HandlePrometheus)
}

func Run() {
	router.Run(":" + fmt.Sprint(config.Conf.Checker.Listen))
}
