package router

import (
	"go-admin/app/sn/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerBoxRelationInfoRouter)
}

// 需认证的路由代码
func registerBoxRelationInfoRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.BoxRelationInfo{}

	r := v1.Group("/box-relation-info")
	{
		r.GET("", api.GetBoxRelationInfoList)
		r.PUT("/:id", api.AddBox)
	}
}
