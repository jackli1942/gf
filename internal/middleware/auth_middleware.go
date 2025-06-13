package middleware

import (
	"context"
	// "strings" // gstr.HasPrefix is better for GoFrame
	"yuncms/internal/service" // For ITokenService and UserTokenClaims

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/text/gstr"
)

// CtxUserClaimsKey is the key for storing user claims in the context.
const CtxUserClaimsKey = "UserClaims"

// AuthMiddleware handles JWT authentication and authorization.
func AuthMiddleware(r *ghttp.Request) {
	ctx := r.Context()

	// 1. Define routes/paths that are exempt from authentication
	// This is a simplified list. A more robust solution would use route metadata or configuration.
	// Example: exemptPaths := g.Cfg().MustGet(ctx, "auth.exemptPaths", g.SliceStr{"/api/v1/auth/login", "/api/v1/health"}).SliceStr()
	exemptPaths := []string{
		"/api/v1/auth/login", // User login
		"/api/v1/health",     // Health check
		// Add other paths like Swagger docs, captcha, etc. if they exist and should be public
	}
	isExempt := false
	currentPath := r.URL.Path
	for _, p := range exemptPaths {
		// Use HasPrefix for paths like /api/v1/health/* or if sub-paths are also exempt.
		// If exact match is needed for some, adjust logic.
		if gstr.HasPrefix(currentPath, p) {
			isExempt = true
			break
		}
	}

	if isExempt {
		r.Middleware.Next()
		return
	}

	// 2. Get token from Authorization header or query parameter
	tokenString := r.Header.Get("Authorization")
	if tokenString == "" {
		tokenString = r.Get("token").String() // Common query param name
	} else {
		// Remove "Bearer " prefix if present
		tokenString = gstr.Replace(tokenString, "Bearer ", "", 1)
	}

	if tokenString == "" {
		r.Response.WriteJsonExit(g.Map{
			"code":    gcode.CodeNotAuthorized.Code(),
			"message": g.I18n().T(ctx, "auth.tokenRequired"), // "Authorization token is required"
		})
		return
	}

	// 3. Parse and validate the JWT structure and signature
	tokenService := service.NewTokenService() // Ideally, this would be from a DI container or a global instance
	claims, err := tokenService.ParseUserToken(ctx, tokenString)
	if err != nil {
		// ParseUserToken already returns gerror with appropriate gcode and i18n-keyed message
		parsedCode := gerror.Code(err)
		r.Response.WriteJsonExit(g.Map{
			"code":    parsedCode.Code(),
			"message": err.Error(), // This will be the i18n message from ParseUserToken
		})
		return
	}

	// 4. Validate token against Redis (is it active, not logged out/kicked?)
	isValidInRedis, redisErr := tokenService.ValidateTokenInRedis(ctx, claims.UserId, tokenString)
	if redisErr != nil {
		g.Log().Errorf(ctx, "AuthMiddleware: Error validating token in Redis for UserID %d: %v", claims.UserId, redisErr)
		r.Response.WriteJsonExit(g.Map{
			"code":    gcode.CodeInternal.Code(), // Or a more specific service unavailable code
			"message": g.I18n().T(ctx, "auth.tokenValidationFailed"), // "Token validation failed due to an internal error"
		})
		return
	}
	if !isValidInRedis {
		// This specific error (e.g., TOKEN_NOT_IN_REDIS, TOKEN_BINDING_MISMATCH) comes from ValidateTokenInRedis
		// We can use its message or a generic one here.
		parsedRedisErrorCode := gerror.Code(redisErr) // Though redisErr might be nil if !isValidInRedis
		if redisErr != nil {
			r.Response.WriteJsonExit(g.Map{
				"code":    parsedRedisErrorCode.Code(),
				"message": redisErr.Error(), // Use the error message from ValidateTokenInRedis
			})
		} else { // Should ideally always have an error if !isValidInRedis
			r.Response.WriteJsonExit(g.Map{
				"code":    gcode.CodeNotAuthorized.Code(),               // Or a more specific "TOKEN_INVALIDATED"
				"message": g.I18n().T(ctx, "auth.tokenInvalidated"), // "Token has been invalidated or logged out"
			})
		}
		return
	}

	// 5. Store user claims in context for subsequent handlers
	// Use r.SetCtx instead of r.SetCtxVar for broader compatibility if context is passed around.
	// r.SetCtxVar(CtxUserClaimsKey, claims) is also fine for ghttp.Request context.
	// Let's use a more common pattern of adding to the context that r.Context() returns.
	newCtx := context.WithValue(r.Context(), CtxUserClaimsKey, claims)
	r.SetCtx(newCtx) // Update the request's context

	r.Middleware.Next()
}

// GetUserClaimsFromCtx retrieves user claims previously stored in the context by AuthMiddleware.
// Returns nil if no claims are found.
func GetUserClaimsFromCtx(ctx context.Context) *service.UserTokenClaims {
	val := ctx.Value(CtxUserClaimsKey)
	if val == nil {
		return nil
	}
	if claims, ok := val.(*service.UserTokenClaims); ok {
		return claims
	}
	return nil
}

// Helper function to get current UserID from context (example of use)
// Returns 0 if claims are not found or UserId is 0.
func GetCurrentUserIdFromCtx(ctx context.Context) uint {
	claims := GetUserClaimsFromCtx(ctx)
	if claims != nil {
		return claims.UserId
	}
	return 0
}

// Helper function to get current Username from context (example of use)
// Returns "" if claims are not found.
func GetCurrentUsernameFromCtx(ctx context.Context) string {
	claims := GetUserClaimsFromCtx(ctx)
	if claims != nil {
		return claims.Username
	}
	return ""
}
```
