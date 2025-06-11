package middleware

import (
	commonLogic "gf_project/internal/logic"                // For common response JsonRes and helpers
	systemLogic "gf_project/internal/modules/system/logic" // For CasbinEnforcer

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/util/gconv"
)

// sCasbinMiddleware provides Casbin authorization middleware.
type sCasbinMiddleware struct{}

// CasbinAuth is the exported Casbin middleware instance.
var CasbinAuth = sCasbinMiddleware{}

// Middleware checks Casbin permissions for the current request.
// It should be used AFTER the JWT authentication middleware (Auth.Middleware).
func (s *sCasbinMiddleware) Middleware(r *ghttp.Request) {
	// 1. Get the Casbin enforcer
	enforcer, err := systemLogic.CasbinEnforcer() // Assumes CasbinEnforcer() is in systemLogic
	if err != nil {
		g.Log().Errorf(r.Context(), "Casbin enforcer not available during permission check: %v", err)
		// Use commonLogic.Respond directly as controller is not involved here
		r.Response.WriteJsonExit(commonLogic.JsonRes{
			Code:    gcode.CodeInternalError.Code(),
			Message: "Access control system error (enforcer unavailable)",
		})
		return
	}

	// 2. Get the subject (sub) - authenticated user ID
	// UserID should be in context from the JWT middleware (Auth.Middleware)
	userIdFromCtx := r.Context().Value(ContextKeyUserId) // Constant defined in auth_middleware.go
	if userIdFromCtx == nil {
		g.Log().Warning(r.Context(), "Casbin middleware: UserID not found in context. Ensure JWT middleware runs first.")
		r.Response.WriteJsonExit(commonLogic.JsonRes{
			Code:    gcode.CodeNotAuthorized.Code(),
			Message: "User identity not found for permission check (missing UserId in Ctx)",
		})
		return
	}
	sub := gconv.String(userIdFromCtx)

	// 3. Get the object (obj) - the resource being accessed (e.g., API path)
	obj := r.URL.Path // This is the full path, e.g., /api/v1/system/user-admin/list

	// 4. Get the action (act) - the HTTP method
	act := r.Method

	// 5. Enforce policy
	g.Log().Debugf(r.Context(), "Casbin Enforce Check: Subject='%s', Object='%s', Action='%s'", sub, obj, act)
	allowed, enforceErr := enforcer.Enforce(sub, obj, act)
	if enforceErr != nil {
		g.Log().Errorf(r.Context(), "Casbin Enforce() error: %v. Sub: %s, Obj: %s, Act: %s", enforceErr, sub, obj, act)
		r.Response.WriteJsonExit(commonLogic.JsonRes{
			Code:    gcode.CodeInternalError.Code(),
			Message: "Error during permission check",
		})
		return
	}

	if allowed {
		g.Log().Debugf(r.Context(), "Casbin Enforce: GRANTED for Subject='%s', Object='%s', Action='%s'", sub, obj, act)
		r.Middleware.Next()
	} else {
		g.Log().Warningf(r.Context(), "Casbin Enforce: DENIED for Subject='%s', Object='%s', Action='%s'", sub, obj, act)
		r.Response.WriteJsonExit(commonLogic.JsonRes{
			Code:    gcode.CodeForbidden.Code(),
			Message: "Access Denied: You do not have permission to perform this action or access this resource.",
		})
		return
	}
}
