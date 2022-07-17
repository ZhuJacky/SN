package router

import (
	"go-admin/app/sn/apis"

	"github.com/gin-gonic/gin"
	jwt "github.com/go-admin-team/go-admin-core/sdk/pkg/jwtauth"
)

func init() {
	routerCheckRole = append(routerCheckRole, registerSnProductRouter)
}

// 需认证的路由代码
func registerSnProductRouter(v1 *gin.RouterGroup, authMiddleware *jwt.GinJWTMiddleware) {
	api := apis.ProductInfo{}
	// r := v1.Group("/sn-batch").Use(authMiddleware.MiddlewareFunc()).Use(middleware.AuthCheckRole())
	// {
	// 	r.GET("", api.GetPage)
	// 	r.GET("/:id", api.Get)
	// 	r.POST("", api.Insert)
	// 	r.PUT("/:id", api.Update)
	// 	r.DELETE("", api.Delete)
	// }

	r := v1.Group("/sn-product")
	{
		r.GET("", api.GetPage)
		r.GET("/:id", api.Get)
		r.POST("", api.Insert)
		r.PUT("/:id", api.Update)
		r.DELETE("", api.Delete)
	}
}
