package router

import (
	"github.com/gogf/gf/v2/net/ghttp"
	appCtrl "yuncms/internal/app/controller" // User controller from internal/app
	"yuncms/internal/controller"             // Department controller from internal/controller
	"yuncms/internal/controller/health"
	"yuncms/internal/middleware" // Import the middleware
)

// BindController registers routes for all controllers and applies middleware.
func BindController(s *ghttp.Server) {
	s.Group("/", func(rootGroup *ghttp.RouterGroup) {
		// Health check is public
		rootGroup.Bind(
			health.New(),
		)

		// API v1 Group with Authentication Middleware
		apiV1Group := rootGroup.Group("/api/v1")
		// Apply AuthMiddleware first, then CasbinMiddleware
		apiV1Group.Middleware(middleware.AuthMiddleware, middleware.CasbinMiddleware)

		apiV1Group.Bind(
			appCtrl.NewUser(),          // Registers routes from UserController in internal/app/controller
			controller.NewDepartmentController(), // Registers routes from DepartmentController in internal/controller
			// Add other protected controllers here
		)

		// Note: If there are specific public routes within /api/v1 (besides /api/v1/auth/login which is handled by exemptions),
		// they would need to be defined outside this group or have the middleware selectively applied/bypassed.
		// For now, all of /api/v1 is protected except explicitly exempted paths in AuthMiddleware.
	})

	// Example: If login was outside /api/v1 group and public
	// s.Group("/", func(publicGroup *ghttp.RouterGroup) {
	//     publicGroup.Bind(
	//         appCtrl.NewUser(), // Assuming Login is part of NewUser and its g.Meta path makes it distinct
	//     )
	// })
}
