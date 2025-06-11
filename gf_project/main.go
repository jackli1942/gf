package main

import (
	"context"

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	// "github.com/gogf/gf/v2/os/gcmd" // Not strictly needed for the simplified main
	"github.com/gogf/gf/v2/os/gctx"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"

	// "gf_project/internal/cmd"      // For potential sub-commands, not used in this simplified main
	"gf_project/internal/logic/middleware" // Ensure this is imported for Auth.Init()
	systemLogic "gf_project/internal/modules/system/logic"
	_ "gf_project/internal/modules/system/router" // Ensure router is imported so its init() runs
	// _ "gf_project/internal/packed"
)

func main() {
	ctx := gctx.New()

	// Initialize Casbin Enforcer
	g.Log().Info(ctx, "Initializing Casbin Enforcer...")
	systemLogic.InitCasbinEnforcer(ctx)

	// Initialize JWT Middleware
	g.Log().Info(ctx, "Initializing JWT Middleware...")
	if err := middleware.Auth.Init(ctx); err != nil {
		g.Log().Fatalf(ctx, "JWT Middleware Init Error: %v", err)
	}

	// Server configuration
	s := g.Server()
	s.Use(middleware.Recovery) // Should be early
	s.Use(middleware.Logger)
	s.Use(middleware.CORS)
	// TODO: Add other global middleware like request tracing, rate limiting etc.

	// Module Routers are initialized due to underscore imports above (e.g., system/router)
	// The routes defined in those init() functions will be registered to g.Server().

	g.Log().Info(ctx, "Starting GoFrame HTTP server...")
	s.Run()
}
