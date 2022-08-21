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

	r1 := v1.Group("/box-info")
	{
		r1.GET("", api.GetBoxInfoList)
		r1.PUT("/:id", api.UpdateBoxSum)
	}

	r2 := v1.Group("/ex-warehouse")
	{
		r2.POST("/do-ex-warehouse", api.UpdateExWarehouseBoxStatus)
		r2.GET("/ex-warehouse-box", api.GetExWarehouseBoxList)
	}
}
