// Package router provides ...
package router

import (
	"fmt"

	"mysslee_qcloud/app/backend/services"
	_ "mysslee_qcloud/app/backend/services/payment"
	"mysslee_qcloud/config"
	"mysslee_qcloud/mylog"
	"mysslee_qcloud/qcloud"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

var router = gin.New()

func init() {
	if config.Conf.RunMode == config.MODE_PRE_PRODUCTION ||
		config.Conf.RunMode == config.MODE_PRODUCTION {
		gin.SetMode(gin.ReleaseMode)

		logrus.SetLevel(logrus.InfoLevel)
	}

	router.Use(mylog.Logger())
	router.Use(gin.Recovery())

	// router
	router.POST("/external/call", func(c *gin.Context) {
		qcloud.HandleExternalCall(c, services.GetAccount)
	})
	router.POST("/internal/call", func(c *gin.Context) {
		qcloud.HandleInternalCall(c)
	})
	internal := router.Group("/api", services.InternalFilter)
	{
		// app info
		internal.GET("/app/infos", services.HandleAppInfos)
		// get account plan
		internal.GET("/plan/:uin", services.HandleAccountPlan)
		// handle kafka check callback
		internal.POST("/taskCallback", services.HandleTaskCallback)
	}
	// prometheus
	//router.GET("/metrics", prom.HandlePrometheus)
}

// Run start gin
func Run() {
	router.Run(":" + fmt.Sprint(config.Conf.Backend.Listen))
}
