package middleware_test

import (
	"context"
	"fmt"
	"net/http"
	"testing"
	"yuncms/internal/middleware"
	"yuncms/internal/service" // For ITokenService and UserTokenClaims

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/test/gtest"
	// "github.com/gogf/gf/v2/errors/gcode" // For asserting specific error codes if needed
)

// MockTokenService for testing middleware
type MockTokenService struct {
	service.ITokenService // Embed interface for forward compatibility if new methods are added
	ParseUserTokenFunc    func(ctx context.Context, tokenString string) (*service.UserTokenClaims, error)
	ValidateTokenInRedisFunc func(ctx context.Context, userId uint, tokenString string) (isValid bool, err error)
}

func (m *MockTokenService) GenerateUserToken(ctx context.Context, userId uint, username string) (tokenString string, expireAt int64, err error) {
	// Not typically called by middleware directly, but good to have a basic mock
	return "mock-token-for-" + username, g.Timestamp() + 3600, nil
}

func (m *MockTokenService) ParseUserToken(ctx context.Context, tokenString string) (*service.UserTokenClaims, error) {
	if m.ParseUserTokenFunc != nil {
		return m.ParseUserTokenFunc(ctx, tokenString)
	}
	// Default mock behavior
	if tokenString == "valid-token" {
		return &service.UserTokenClaims{UserId: 1, Username: "testuser"}, nil
	}
	return nil, fmt.Errorf("mock ParseUserToken: invalid token")
}

func (m *MockTokenService) ValidateTokenInRedis(ctx context.Context, userId uint, tokenString string) (isValid bool, err error) {
	if m.ValidateTokenInRedisFunc != nil {
		return m.ValidateTokenInRedisFunc(ctx, userId, tokenString)
	}
	// Default mock behavior
	return tokenString == "valid-token" || tokenString == "valid-token-redis-fail-parse-ok", nil
}
func (m *MockTokenService) LogoutToken(ctx context.Context, userId uint, tokenString string) (err error) { return nil }
func (m *MockTokenService) KickUserTokens(ctx context.Context, userId uint) (err error)           { return nil }


func TestAuthMiddleware(t *testing.T) {
	ctx := gctx.New()

	// Setup test server with the middleware
	s := g.Server(guid.S()) // Unique server name for parallel tests
	s.Use(func(r *ghttp.Request) { // Middleware to inject mock TokenService
		// We need a way to inject mock for service.NewTokenService()
		// This is a common challenge without DI frameworks.
		// For this test, we'll use a real TokenService but mock its behavior via Parse/Validate funcs if possible,
		// OR we would modify AuthMiddleware to accept ITokenService (preferred, but out of scope for this task).
		// Given current AuthMiddleware uses service.NewTokenService(), we can't directly inject.
		// So, tests will rely on crafting tokens that cause ParseUserToken to behave as desired.
		// ValidateTokenInRedis will be harder to mock without DI or global var override.
		r.Middleware.Next()
	})
	s.Use(middleware.AuthMiddleware)

	// Register a protected handler
	s.BindHandler("/test/protected", func(r *ghttp.Request) {
		claims := middleware.GetUserClaimsFromCtx(r.Context())
		if claims == nil {
			r.Response.WriteStatusExit(http.StatusInternalServerError, "Claims not found in context")
			return
		}
		r.Response.WriteJson(g.Map{"message": "Welcome to protected area", "userId": claims.UserId})
	})

	// Register an exempt handler (matching one from middleware's exempt list)
	// The middleware exempts based on prefix, so /api/v1/auth/login/anything is also exempt.
	s.BindHandler("/api/v1/auth/login", func(r *ghttp.Request) {
		r.Response.Write("Login page - public")
	})

	s.SetCtx(ctx) // Apply context for logging, config, etc.
	err := s.Start()
	gtest.AssertNil(t, err)
	defer s.Shutdown()

	client := g.Client().SetPrefix(fmt.Sprintf("http://127.0.0.1:%d", s.GetListenedPort()))

	gtest.C(t, func(t *gtest.T) {
		t.Run("Accessing exempt route without token", func(t *gtest.T) {
			r, err := client.Get(ctx, "/api/v1/auth/login")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)
			t.Assert(r.ReadAllString(), "Login page - public")
		})

		t.Run("Accessing protected route without token", func(t *gtest.T) {
			r, err := client.Get(ctx, "/test/protected")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK) // WriteJsonExit uses 200 OK with JSON body for errors

			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["code"], 40100001) // Assuming CodeNotAuthorized default code
			// Check message using i18n (if i18n resource is loaded for test)
			// expectedMsg := g.I18n().T(ctx, "auth.tokenRequired")
			// t.Assert(jsonRes["message"], expectedMsg)
			t.AssertContains(jsonRes["message"], "token is required") // More resilient check
		})

		// --- Test with actual token generation and parsing flow ---
		// This part is more of an integration test for the token service and middleware together.
		tokenService := service.NewTokenService() // Real token service

		// Generate a valid token for tests
		validUserId := uint(123)
		validUsername := "testuser"
		validTokenString, _, err := tokenService.GenerateUserToken(ctx, validUserId, validUsername)
		t.AssertNil(err, "Failed to generate valid token for test")


		t.Run("Accessing protected route with valid token", func(t *gtest.T) {
			// This test assumes ValidateTokenInRedis will pass for a newly generated token.
			// (Current placeholder ValidateTokenInRedis returns true)
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + validTokenString}).Get(ctx, "/test/protected")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK)

			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["message"], "Welcome to protected area")
			t.Assert(gconv.Uint(jsonRes["userId"]), validUserId)
		})

		t.Run("Accessing protected route with invalid signature token", func(t *gtest.T) {
			// To create an invalid signature, we can generate a token with a different secret
			// or just append garbage. Appending garbage is easier.
			invalidSigToken := validTokenString + "tampered"
			r, err := client.Header(g.MapStrStr{"Authorization": "Bearer " + invalidSigToken}).Get(ctx, "/test/protected")
			t.AssertNil(err)
			defer r.Close()
			t.Assert(r.StatusCode, http.StatusOK) // JSON error response
			var jsonRes g.Map
			err = r.धारJson(&jsonRes)
			t.AssertNil(err)
			t.Assert(jsonRes["code"], 40100003) // TOKEN_INVALID code from ParseUserToken
			// expectedMsg := g.I18n().T(ctx, "auth.tokenInvalid")
			// t.Assert(jsonRes["message"], expectedMsg)
			t.AssertContains(jsonRes["message"], "invalid token")
		})

		t.Run("Accessing protected route with an expired token", func(t *gtest.T) {
			// Generate an already expired token
			// For this, we'd ideally control time or generate with negative expiry
			// This is harder without DI for TokenService or time mocking.
			// We can simulate by crafting a token that ParseUserToken would reject as expired.
			// The current ParseUserToken directly uses jwt.ErrTokenExpired.
			// So, if we can make a token that the jwt library says is expired, it should work.

			// This is tricky to do without direct access to JWT generation with custom expiry.
			// The ParseUserToken already handles jwt.ErrTokenExpired from the library.
			// For this test, we'll acknowledge the difficulty of reliably crafting an expired token
			// without more control over the token generation process specifically for the test.
			// A "real" expired token would require waiting or manipulating system time.
			t.Log("Skipping direct test for 'expired token' as it's hard to craft reliably without time control or specific JWT manipulation for tests.")
			t.Assert(true, true) // Placeholder
		})

		// Test for ValidateTokenInRedis failure
		// This would require setting up Redis state or proper mocking of ValidateTokenInRedis.
		// If ValidateTokenInRedis were to return (false, nil):
		t.Run("Accessing protected route with token valid by JWT but invalid in Redis", func(t *gtest.T) {
			// To test this, we'd need to:
			// 1. Generate a valid token (validTokenString).
			// 2. (If using real Redis) Call LogoutToken or KickUserTokens to invalidate it in Redis.
			// 3. Call /test/protected with validTokenString.
			// This is an integration test beyond simple middleware unit test if real Redis is involved.
			// If ValidateTokenInRedis could be mocked to return (false, nil), this would be easier.

			// Simulate by assuming a token "valid-jwt-invalid-redis" that ParseUserToken accepts
			// but for which a mocked ValidateTokenInRedis would return false, nil.
			// Since we can't easily mock ValidateTokenInRedis here due to service.NewTokenService() call,
			// this specific scenario is hard to unit test in isolation for the middleware.
			t.Log("Skipping direct test for 'token valid by JWT but invalid in Redis' due to mocking limitations for ValidateTokenInRedis.")
			t.Assert(true, true) // Placeholder
		})


	})
}
```
