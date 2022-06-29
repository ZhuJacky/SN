// Package main provides ...
package main

import (
	"github.com/gin-gonic/gin"

	"mysslee_qcloud/test/billing"
	"mysslee_qcloud/test/proxy"
)

func main() {
	router := gin.Default()

	router.POST("/external/api", proxy.ProxyExternal)
	router.POST("/internal/api", proxy.ProxyInternal)

	// BillingRoute
	router.GET("/order/check", billing.HandleOrderCheck)

	router.Run(":9030")
}
