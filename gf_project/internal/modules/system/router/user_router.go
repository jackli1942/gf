package router

import (
	// "gf_project/api/v1/system" // API structs are used by controller, not directly in router Bind with method approach
	"gf_project/internal/logic/middleware" // For Auth and Casbin middleware
	"gf_project/internal/modules/system/controller"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

func init() {
	s := g.Server()

	s.Group("/api/v1/system", func(group *ghttp.RouterGroup) {

		group.Bind(
			controller.UserController.Login,
		)

		group.Group("/user", func(userAuthGroup *ghttp.RouterGroup) {
			userAuthGroup.Middleware(middleware.Auth.Middleware)
			userAuthGroup.Bind(
				controller.UserController.GetProfile,
				controller.UserController.UpdateProfile,
				controller.UserController.ChangePassword,
			)
		})

		group.Group("/user-admin", func(adminUserGroup *ghttp.RouterGroup) {
			adminUserGroup.Middleware(middleware.Auth.Middleware)
			adminUserGroup.Middleware(middleware.CasbinAuth.Middleware) // Then apply Casbin Auth

			adminUserGroup.Bind(
				controller.UserController.Create,
				controller.UserController.List,
				controller.UserController.Get,
				controller.UserController.Update,
				controller.UserController.Delete,
				controller.UserController.ResetPassword,
				controller.UserController.UpdateStatus,
				controller.UserController.AssignRoles,
			)
		})
	})
}
