package router

import (
	// "gf_project/api/v1/system" // API structs are used by controller, not directly in router Bind with method approach
	"gf_project/internal/logic/middleware"          // For Auth and Casbin middleware
	"gf_project/internal/modules/system/controller" // For MenuController

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// init registers the routes for the system menu management module.
func init() {
	s := g.Server()

	s.Group("/api/v1/system", func(group *ghttp.RouterGroup) {

		// User-specific menu route (e.g., fetching their own accessible menus)
		// This route requires only JWT authentication.
		// Authorization (which menus are returned) is handled by the service logic (GetUserMenus).
		group.Group("/menu", func(menuUserGroup *ghttp.RouterGroup) {
			menuUserGroup.Middleware(middleware.Auth.Middleware) // Apply JWT Auth
			menuUserGroup.Bind(
				controller.MenuController.UserMenus,
			)
		})

		// Admin actions for Menu Management
		// These routes require JWT authentication and Casbin permission checks.
		group.Group("/menu-admin", func(menuAdminGroup *ghttp.RouterGroup) {
			menuAdminGroup.Middleware(middleware.Auth.Middleware)       // 1. Apply JWT Authentication Middleware
			menuAdminGroup.Middleware(middleware.CasbinAuth.Middleware) // 2. Apply Casbin Authorization Middleware

			// Bind controller methods. GoFrame automatically maps API request structs
			// (defined in api/v1/system/menu.go with g.Meta tags) to these methods.
			menuAdminGroup.Bind(
				controller.MenuController.Create,
				controller.MenuController.List,
				controller.MenuController.Get,
				controller.MenuController.Update,
				controller.MenuController.Delete,
			)
		})
	})
}
