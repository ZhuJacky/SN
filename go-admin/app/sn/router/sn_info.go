package router

import (
	"go-admin/app/sn/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerSNInfoRouter)
}

// 需认证的路由代码
func registerSNInfoRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.SNInfo{}

	r := v1.Group("/sn-info")
	{
		r.GET("", api.GetSNInfoList)
		r.GET("/:id", api.Get)
		r.PUT("/:id", api.UpdateStatus)

		// r.POST("/packbox", api.PackBox)
	}
}
