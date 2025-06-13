package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"yuncms/internal/app/service" // For service.Casbin()
	"yuncms/internal/middleware"
	gfService "yuncms/internal/service" // For ITokenService and UserTokenClaims

	"github.com/casbin/casbin/v2"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/util/gconv"
	// "github.com/gogf/gf/v2/errors/gcode" // For asserting specific error codes
)

// setupCasbinEnforcerForTest clears existing policies and loads test-specific ones.
// IMPORTANT: This manipulates the global/singleton enforcer returned by service.Casbin().
// This is suitable for testing environments where tests run serially or
// if the enforcer is re-initialized between test packages/runs.
func setupCasbinEnforcerForTest(t *gtest.T, e *casbin.Enforcer, policies [][]string, groupPolicies [][]string) {
	e.ClearPolicy() // Clear all policies (p) and groups (g)

	if len(policies) > 0 {
		added, err := e.AddPolicies(policies)
		gtest.AssertNil(t, err, "Error adding test policies to Casbin enforcer")
		gtest.Assert(t, added, true, "Failed to add test policies to Casbin enforcer")
	}

	if len(groupPolicies) > 0 {
		added, err := e.AddGroupingPolicies(groupPolicies)
		gtest.AssertNil(t, err, "Error adding test grouping policies to Casbin enforcer")
		gtest.Assert(t, added, true, "Failed to add test grouping policies to Casbin enforcer")
	}

	err := e.LoadPolicy() // Reload to ensure changes are effective if using watchers or persistent adapters. For in-memory, this might not be strictly needed but good for consistency.
	gtest.AssertNil(t, err, "Error reloading Casbin policy after test setup")
}


func TestCasbinMiddleware(t *testing.T) {
	ctx := gctx.New()

	// Get the shared Casbin enforcer instance
	// IMPORTANT: Test cases will modify this shared enforcer's policies.
	// This means tests that run in parallel and modify Casbin rules could interfere.
	// gtest runs t.Run blocks serially by default within a single T, which helps.
	enforcer := service.Casbin()
	if enforcer == nil {
		t.Fatal("Casbin enforcer is nil, cannot proceed with Casbin middleware tests. Ensure Casbin service is initialized.")
	}

	// Token service for generating test tokens
	tokenService := gfService.NewTokenService()

	// Setup test server
	s := g.Server(guid.S())
	s.SetCtx(ctx) // For logging, config, i18n inside middleware/handlers

	// Apply AuthMiddleware first, then CasbinMiddleware
	s.Group("/test-casbin", func(group *ghttp.RouterGroup) {
		group.Middleware(middleware.AuthMiddleware, middleware.CasbinMiddleware)
		group.ALL("/resource1", func(r *ghttp.Request) {
			claims := middleware.GetUserClaimsFromCtx(r.Context())
			r.Response.WriteJson(g.Map{"message": "Access Granted to resource1", "userId": claims.UserId})
		})
		group.GET("/resource2", func(r *ghttp.Request) {
			claims := middleware.GetUserClaimsFromCtx(r.Context())
			r.Response.WriteJson(g.Map{"message": "Access Granted to resource2", "userId": claims.UserId})
		})
	})

	// Public route, not under /test-casbin, so CasbinMiddleware won't apply
	s.GET("/public/ping", func(r *ghttp.Request) {
		r.Response.Write("pong")
	})

	// Login route (exempted by AuthMiddleware, so CasbinMiddleware won't run)
	// For this test, we just need a path that AuthMiddleware would exempt
	s.POST("/api/v1/auth/login", func(r *ghttp.Request) {
		r.Response.Write("mock login success")
	})


	err := s.Start()
	gtest.AssertNil(t, err)
	defer s.Shutdown()

	client := g.Client().SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort()))

	gtest.C(t, func(t *gtest.T) {
		// Define test users
		user123ID := uint(123)
		user123Token, _, _ := tokenService.GenerateUserToken(ctx, user123ID, "user123")

		user456ID := uint(456)
		user456Token, _, _ := tokenService.GenerateUserToken(ctx, user456ID, "user456")

		user789ID := uint(789)
		user789Token, _, _ := tokenService.GenerateUserToken(ctx, user789ID, "user789")

		// --- Test Policies Set 1 ---
		setupCasbinEnforcerForTest(t, enforcer,
			[][]string{
				{"user_123", "/test-casbin/resource1", "GET"},    // User 123 can GET resource1
				{"role_editor", "/test-casbin/resource1", "POST"}, // Editors can POST resource1
				{"role_viewer", "/test-casbin/resource2", "GET"},  // Viewers can GET resource2
			},
			[][]string{
				{gconv.String(user456ID), "role_editor"}, // User 456 is an editor
				{gconv.String(user123ID), "role_viewer"}, // User 123 is also a viewer
			},
		)

		t.Run("User with Direct Permission (user_123 GET /resource1)", func(t *gtest.T) {
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + user123Token}).Get(ctx, "/test-casbin/resource1")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["message"], "Access Granted to resource1")
			t.Assert(gconv.Uint(jsonRes["userId"]), user123ID)
		})

		t.Run("User DENIED Direct Permission (user_123 POST /resource1)", func(t *gtest.T) {
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + user123Token}).Post(ctx, "/test-casbin/resource1", nil)
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK) // Middleware WriteJsonExit uses 200 OK
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["code"], 40300000) // gcode.CodeForbidden
			// t.Assert(jsonRes["message"], g.I18n().T(ctx, "auth.permissionDenied")) // Requires i18n loaded
			t.AssertContains(jsonRes["message"], "Permission Denied")
		})

		t.Run("User with Role-Based Permission (user_456 POST /resource1 via role_editor)", func(t *gtest.T) {
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + user456Token}).Post(ctx, "/test-casbin/resource1", nil)
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["message"], "Access Granted to resource1")
			t.Assert(gconv.Uint(jsonRes["userId"]), user456ID)
		})

		t.Run("User DENIED Role-Based Permission (user_456 GET /resource1)", func(t *gtest.T) {
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + user456Token}).Get(ctx, "/test-casbin/resource1")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["code"], 40300000)
		})

		t.Run("User with another Role-Based Permission (user_123 GET /resource2 via role_viewer)", func(t *gtest.T) {
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + user123Token}).Get(ctx, "/test-casbin/resource2")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["message"], "Access Granted to resource2")
		})


		t.Run("User with No Relevant Permissions (user_789 GET /resource1)", func(t *gtest.T) {
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + user789Token}).Get(ctx, "/test-casbin/resource1")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["code"], 40300000)
		})

		t.Run("Access to protected route without Token (caught by AuthMiddleware)", func(t *gtest.T) {
			r, err := client.Get(ctx, "/test-casbin/resource1") // No token
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["code"], 40100001) // CodeNotAuthorized from AuthMiddleware
			t.AssertContains(jsonRes["message"], "token is required")
		})

		t.Run("Access to public route (not under Casbin group)", func(t *gtest.T) {
			r, err := client.Get(ctx, "/public/ping")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			t.Assert(r.ReadAllString(), "pong")
		})

		t.Run("Access to login route (exempt by AuthMiddleware)", func(t *gtest.T) {
			r, err := client.Post(ctx, "/api/v1/auth/login", nil) // No token needed
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			t.Assert(r.ReadAllString(), "mock login success")
		})

	})

	// Restore original policies if necessary, or ensure tests don't affect global state if run in parallel.
	// For this structure, ClearPolicy at the start of setupCasbinEnforcerForTest is key.
}
```
