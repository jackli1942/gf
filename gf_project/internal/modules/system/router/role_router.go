package router

import (
	// "gf_project/api/v1/system" // API structs are used by controller, not directly in router Bind with method approach
	"gf_project/internal/logic/middleware"          // For Auth and Casbin middleware
	"gf_project/internal/modules/system/controller" // For RoleController

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
)

// init registers the routes for the system role management module.
// This function is automatically called by GoFrame during server initialization
// because this package (router) is imported (likely via an underscore import in main.go
// or a module bootstrap file that imports all routers in this directory).
func init() {
	s := g.Server()

	// All role management routes are admin-only and require auth + specific permissions.
	// They are grouped under /api/v1/system, similar to user admin routes.
	s.Group("/api/v1/system", func(group *ghttp.RouterGroup) {
		// Define a sub-group specifically for role administration.
		// This ensures that middleware applied here is specific to role admin actions.
		group.Group("/role-admin", func(roleAdminGroup *ghttp.RouterGroup) {
			roleAdminGroup.Middleware(middleware.Auth.Middleware)       // 1. Apply JWT Authentication Middleware
			roleAdminGroup.Middleware(middleware.CasbinAuth.Middleware) // 2. Apply Casbin Authorization Middleware

			// Bind controller methods. GoFrame will automatically map API request structs
			// (defined in api/v1/system/role.go with g.Meta tags) to these methods
			// because the controller methods have the signature:
			// func(ctx context.Context, req *system.XxxReq) (*system.XxxRes, error)
			// and the method names (e.g., Create, List) match the OperationID or method name
			// conventions often derived from the API struct name (e.g., RoleCreateReq -> Create method).
			roleAdminGroup.Bind(
				controller.RoleController.Create,
				controller.RoleController.List,
				controller.RoleController.Get,
				controller.RoleController.Update,
				controller.RoleController.Delete,
				controller.RoleController.UpdatePermissions,
				controller.RoleController.GetPermissions,
				controller.RoleController.UpdateStatus,
			)
		})
	})
}
