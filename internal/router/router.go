package router

import (
	"github.com/gogf/gf/v2/net/ghttp"
	"yuncms/internal/controller/health"
	// Import other controllers here as they are created
)

// BindController registers routes for all controllers.
// It's a common pattern to have a function that groups all router registrations.
func BindController(s *ghttp.Server) {
	s.Group("/", func(group *ghttp.RouterGroup) {
		// Register health check controller
		group.Bind(
			health.New(),
		)

		// Example of a simple root handler (can be removed if health controller is at root path or if other groups are defined)
		// group.ALL("/", func(r *ghttp.Request) {
		//	r.Response.Writeln("Welcome to Yuncms API")
		// })

		// TODO: Register other domain controllers like user, department, etc.
		// group.Group("/api/v1", func(apiV1Group *ghttp.RouterGroup) {
		// apiV1Group.Middleware(service.Middleware().Auth) // Example middleware
		// apiV1Group.Bind(
		// user.NewV1(),
		// ... other v1 controllers
		// )
		// })
	})
}
