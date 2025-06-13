package middleware

import (
	"yuncms/internal/app/service" // For casbin_service.Casbin()

	// For GetUserClaimsFromCtx, assuming it's in the same package or made accessible.
	// If it were in, for example, "yuncms/internal/service", it would be:
	// tokenService "yuncms/internal/service"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

// CasbinMiddleware handles Casbin-based authorization.
// It should run AFTER AuthMiddleware.
func CasbinMiddleware(r *ghttp.Request) {
	ctx := r.Context()

	// 1. Get authenticated user claims from context (populated by AuthMiddleware)
	claims := GetUserClaimsFromCtx(ctx) // Assumes this is in the same 'middleware' package or otherwise accessible

	if claims == nil || claims.UserId == 0 {
		// This scenario implies that AuthMiddleware might have passed the request through
		// (e.g., for an exempt path from JWT check, or a misconfiguration),
		// or the claims were not set.
		// If a path reaches CasbinMiddleware, it typically means it requires authorization.
		g.Log().Warning(ctx, "CasbinMiddleware: No user claims found in context. Ensure AuthMiddleware runs first and correctly sets claims for protected routes.")
		r.Response.WriteJsonExit(g.Map{
			"code":    gcode.CodeNotAuthorized.Code(), // Or perhaps CodeForbidden if auth passed but claims are missing
			"message": g.I18n().T(ctx, "auth.accessDenied"), // "Access Denied: Valid authentication required for authorization."
		})
		return
	}

	// 2. Prepare Casbin enforce parameters
	// Subject: User ID (as string). Casbin policies can then map this user ID to roles or directly to permissions.
	// Example Casbin model:
	// [request_definition]
	// r = sub, obj, act
	// [policy_definition]
	// p = sub, obj, act
	// [role_definition]
	// g = _, _
	// [policy_effect]
	// e = some(where (p.eft == allow))
	// [matchers]
	// m = g(r.sub, p.sub) && r.obj == p.obj && r.act == p.act
	// Policies:
	// p, admin_role, /api/v1/users, GET
	// p, user_123, /api/v1/profile, GET
	// Grouping:
	// g, user_123, admin_role  (User 123 has admin_role)

	subject := gconv.String(claims.UserId) // Using UserID as the subject for Casbin
	object := r.URL.Path                   // The resource path being accessed
	action := r.Method                     // The HTTP method (GET, POST, PUT, DELETE)

	// 3. Perform Casbin check
	casbinEnforcer := service.Casbin() // Get the Casbin enforcer from yuncms/internal/app/service/casbin_service.go
	if casbinEnforcer == nil {
		g.Log().Error(ctx, "CasbinMiddleware: Casbin enforcer is not initialized.")
		r.Response.WriteJsonExit(g.Map{
			"code":    gcode.CodeInternal.Code(),
			"message": g.I18n().T(ctx, "auth.permissionCheckFailed"), // "Permission check failed due to an internal error."
		})
		return
	}

	hasPermission, err := casbinEnforcer.Enforce(subject, object, action)
	if err != nil {
		g.Log().Errorf(ctx, "CasbinMiddleware: Casbin Enforce error for sub:'%s' obj:'%s' act:'%s': %v", subject, object, action, err)
		r.Response.WriteJsonExit(g.Map{
			"code":    gcode.CodeInternal.Code(),
			"message": g.I18n().T(ctx, "auth.permissionCheckFailed"),
		})
		return
	}

	if !hasPermission {
		g.Log().Noticef(ctx, "CasbinMiddleware: Permission denied for subject '%s' on object '%s' action '%s'", subject, object, action)
		r.Response.WriteJsonExit(g.Map{
			"code":    gcode.CodeForbidden.Code(),
			"message": g.I18n().T(ctx, "auth.permissionDenied"), // "Permission Denied."
		})
		return
	}

	// 4. If permission granted, proceed to next handler
	g.Log().Debugf(ctx, "CasbinMiddleware: Permission GRANTED for subject '%s' on object '%s' action '%s'", subject, object, action)
	r.Middleware.Next()
}
```
