package router

import (
	"go-admin/app/sn/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerBoxInfoRouter)
}

// 需认证的路由代码
func registerBoxInfoRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.BoxInfo{}

	r := v1.Group("/box-info")
	{
		r.GET("", api.GetBoxInfoList)
		r.PUT("/:id", api.UpdateBoxSum)
	}
}
