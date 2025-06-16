package router

import (
	"yuncms/internal/controller" // Ensure this path is correct
	"github.com/gogf/gf/v2/net/ghttp"
)

func RegisterUserRoutes(group *ghttp.RouterGroup) {
	userCtl := controller.NewUserController() // Corrected variable name
	group.Group("/users", func(userGroup *ghttp.RouterGroup) { // Changed 'group' to 'userGroup' for clarity
		userGroup.POST("/", userCtl.Create)
		userGroup.GET("/{id}", userCtl.GetById)
		userGroup.PUT("/{id}", userCtl.Update)
		userGroup.DELETE("/{id}", userCtl.Delete)
		userGroup.GET("/", userCtl.List)
	})
}
