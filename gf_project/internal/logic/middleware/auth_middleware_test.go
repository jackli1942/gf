package middleware

import (
	"context"
	"fmt"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	systemLogic "gf_project/internal/modules/system/logic"
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input" // For creating test users

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang-jwt/jwt/v4"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	_ "github.com/hailaz/gf-casbin-adapter/v2" // Ensure casbin adapter is linked if CasbinEnforcer is called
)

// TestMain for auth_middleware_test.
func TestMain(m *testing.M) {
	ctx := gctx.GetGlobal()

	projectRoot := gfile.normalize(gfile.Pwd() + "/../../../") // from internal/logic/middleware to project root

	configDirPath := gfile.Join(projectRoot, "manifest", "config")
	configFileActualPath := gfile.Join(configDirPath, "config.yaml")

	if gfile.Exists(configFileActualPath) {
		if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			adapter.SetPath(configDirPath) // Tell default cfg where to find files
			g.Log().Debugf(ctx, "TestMain (auth_middleware_test): Set config path to: %s", configDirPath)
		} else {
			g.Log().Warningf(ctx, "TestMain (auth_middleware_test): Default config adapter is not *gcfg.AdapterFile, cannot SetPath.")
		}
	} else {
		g.Log().Warningf(ctx, "TestMain (auth_middleware_test): config.yaml not found at: %s. JWT defaults will be used.", configFileActualPath)
	}

	// Initialize Auth middleware (loads config from g.Cfg())
	if err := Auth.Init(ctx); err != nil {
		g.Log().Fatalf(ctx, "TestMain: Failed to initialize Auth middleware: %v", err)
	}

	// Initialize Casbin Enforcer as UserService methods might be called which could touch Casbin
	// This also ensures its config (model path) is resolved correctly.
	modelPathFromConfig := g.Cfg().MustGet(ctx, "casbin.modelPath", "manifest/config/casbin_model.conf").String()
	resolvedModelPath := modelPathFromConfig
	if !gfile.IsAbs(modelPathFromConfig) {
		if gfile.Exists(gfile.Join(configDirPath, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(configDirPath, modelPathFromConfig)
		} else if gfile.Exists(gfile.Join(projectRoot, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(projectRoot, modelPathFromConfig)
		}
	}
	if !gfile.Exists(resolvedModelPath) {
		g.Log().Fatalf(ctx, "TestMain: Casbin model file '%s' not found for Casbin init.", resolvedModelPath)
	}
	g.Cfg().Set("casbin.modelPath", resolvedModelPath) // Override with resolved path for service

	if _, err := systemLogic.CasbinEnforcer(); err != nil {
		g.Log().Warningf(ctx, "TestMain: Casbin enforcer failed to initialize (may impact tests that rely on services calling Casbin): %v", err)
	}

	exitCode := m.Run()
	os.Exit(exitCode)
}

// Helper to clear test data from specified tables.
func clearTestAuthData(ctx context.Context, tables ...string) {
	if len(tables) == 0 {
		tables = []string{systemLogic.UserService.Dao.SystemUser.Table()} // Default to user table
	}
	for _, table := range tables {
		// Clear table and reset auto-increment for sqlite if applicable
		_, err := g.DB().Ctx(ctx).Delete(table)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to clear table %s: %v", table, err)
		}
		if g.DB("default").GetConfig().DriverName == "sqlite" {
			// Reset autoincrement sequence for SQLite
			// For other DBs, TRUNCATE might be an option but is often more privileged.
			// This ensures IDs start fresh for tests if needed.
			g.DB().Ctx(ctx).Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", table))
		}
	}
}

func TestAuth_GenerateToken(t *testing.T) {
	ctx := gctx.New()
	gtest.C(t, func(t *gtest.T) {
		testUserID := uint64(999)
		userRoles := []string{"user", "editor"}

		tokenString, expireTime, err := Auth.GenerateToken(ctx, testUserID, userRoles)
		t.AssertNil(err, "GenerateToken should not return error")
		t.AssertNE(tokenString, "", "Generated token string should not be empty")
		t.Assert(!expireTime.IsZero(), "Expire time should not be zero")

		claims := &CustomClaims{}
		parsedToken, errParse := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
			return Auth.Key, nil
		})
		t.AssertNil(errParse, "Parsing generated token should not error")
		t.Assert(parsedToken.Valid, "Generated token should be valid")
		t.AssertEQ(claims.UserID, testUserID, "UserID in claims should match")
		t.AssertEQ(len(claims.Roles), len(userRoles), "Roles count in claims should match")
		// Could add more specific role content check if needed
	})
}

func TestAuth_Middleware_ValidToken(t *testing.T) {
	ctx := gctx.New()
	gtest.C(t, func(t *gtest.T) {
		clearTestAuthData(ctx, systemLogic.UserService.Dao.SystemUser.Table()) // Clean before test

		uniqueSuffix := gtime.TimestampMilliStr() // More unique suffix
		testUserCreateInp := &input.UserCreateInp{
			Username: "mid_user_" + uniqueSuffix, Password: "password123", Nickname: "MiddlewareValidUser", Status: 0,
		}
		createdUserId, errUser := systemLogic.UserService.CreateUser(ctx, testUserCreateInp)
		t.AssertNil(errUser, "Test user creation should not fail for middleware test")
		t.AssertGT(createdUserId, 0, "Test user ID should be positive")

		userRoles := []string{"test_role"}
		tokenString, _, _ := Auth.GenerateToken(ctx, uint64(createdUserId), userRoles)

		r := httptest.NewRequest("GET", "/api/v1/system/user/profile", nil) // Use a realistic path
		r.Header.Set("Authorization", Auth.TokenHeadName+" "+tokenString)
		w := httptest.NewRecorder()

		server := g.Server() // Use a temporary server for routing context
		server.SetCtx(ctx)
		gfReq := ghttp.NewRequest(server, r, w) // Create GoFrame request

		nextCalled := false
		dummyHandler := func(r *ghttp.Request) {
			nextCalled = true
			ctxUserId := r.Context().Value(ContextKeyUserId)
			t.AssertNE(ctxUserId, nil, "ContextKeyUserId should be set by middleware")
			t.AssertEQ(gconv.Uint64(ctxUserId), uint64(createdUserId), "ContextKeyUserId should match token's UserID")

			ctxUserObj := r.Context().Value(ContextKeyUserObj)
			t.AssertNE(ctxUserObj, nil, "ContextKeyUserObj should be set")
			if u, ok := ctxUserObj.(*entity.SystemUser); ok {
				t.AssertEQ(u.Id, uint64(createdUserId), "User object in context should have correct ID")
			} else {
				t.Errorf("ContextKeyUserObj is not of type *entity.SystemUser, got %T", ctxUserObj)
			}

			ctxUserRoles := r.Context().Value(ContextKeyUserRoles)
			t.AssertNE(ctxUserRoles, nil, "ContextKeyUserRoles should be set")
			if roles, ok := ctxUserRoles.([]string); ok {
				t.AssertEQ(len(roles), 1, "Roles count in context should be 1")
				t.Assert(roles[0] == "test_role", "Role in context should be 'test_role'")
			} else {
				t.Errorf("ContextKeyUserRoles is not []string, got %T", ctxUserRoles)
			}
			r.Response.Write("next handler called") // Indicate next was called
		}

		// Simulate a route and serve it to test middleware in context
		server.BindHandler("/api/v1/system/user/profile", Auth.Middleware(dummyHandler))
		server.ServeHTTP(w, r)

		t.Assert(nextCalled, "Next handler should be called for a valid token")
		t.Assert(w.Code == 200 || w.Code == 0, fmt.Sprintf("Response code should be 200 or 0 (if no explicit write by dummy), got %d", w.Code))

		clearTestAuthData(ctx, systemLogic.UserService.Dao.SystemUser.Table()) // Clean up created user
	})
}

func TestAuth_Middleware_InvalidToken(t *testing.T) {
	ctx := gctx.New()
	gtest.C(t, func(t *gtest.T) {
		r := httptest.NewRequest("GET", "/api/v1/some/protected/route_invalid", nil)
		r.Header.Set("Authorization", Auth.TokenHeadName+" "+"this.is.an.invalid.token")
		w := httptest.NewRecorder()

		server := g.Server()
		server.SetCtx(ctx)
		gfReq := ghttp.NewRequest(server, r, w)

		nextCalled := false
		dummyHandler := func(r *ghttp.Request) { nextCalled = true }

		Auth.Middleware(gfReq) // Call middleware directly for this test structure
		// If it calls WriteJsonExit, it won't proceed to dummyHandler if using server.ServeHTTP for this one.
		// The direct call to Auth.Middleware will have its Respond call r.Exit(), so nextCalled remains false.

		t.Assert(w.Code == gcode.CodeNotAuthorized.Code(),
			fmt.Sprintf("Response code should be NotAuthorized for invalid token, got %d", w.Code))
		t.Assert(!nextCalled, "Next handler should NOT be called for an invalid token")
	})
}

func TestAuth_Middleware_MissingHeader(t *testing.T) {
	ctx := gctx.New()
	gtest.C(t, func(t *gtest.T) {
		r := httptest.NewRequest("GET", "/api/v1/some/protected/route_missing", nil) // No Authorization header
		w := httptest.NewRecorder()

		server := g.Server()
		server.SetCtx(ctx)
		gfReq := ghttp.NewRequest(server, r, w)

		nextCalled := false
		dummyHandler := func(r *ghttp.Request) { nextCalled = true }

		Auth.Middleware(gfReq)

		t.Assert(w.Code == gcode.CodeNotAuthorized.Code(),
			fmt.Sprintf("Response code should be NotAuthorized for missing header, got %d", w.Code))
		t.Assert(!nextCalled, "Next handler should NOT be called for missing header")
	})
}
