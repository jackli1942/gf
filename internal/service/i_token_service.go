package service

import (
	"context"
)

// UserTokenClaims holds the custom claims for our JWT.
type UserTokenClaims struct {
	UserId   uint   `json:"uid"`
	Username string `json:"una"`
	// Add other claims like roles if needed in future
}

// ITokenService defines the interface for JWT and token lifecycle management.
type ITokenService interface {
	GenerateUserToken(ctx context.Context, userId uint, username string) (tokenString string, expireAt int64, err error)
	ParseUserToken(ctx context.Context, tokenString string) (claims *UserTokenClaims, err error)
	// ValidateTokenInRedis will be implemented later to check against Redis store
	ValidateTokenInRedis(ctx context.Context, userId uint, tokenString string) (isValid bool, err error)
	// LogoutToken will be implemented later to invalidate token in Redis
	LogoutToken(ctx context.Context, userId uint, tokenString string) (err error)
	// KickUserTokens will be implemented later to invalidate all tokens for a user
	KickUserTokens(ctx context.Context, userId uint) (err error)
}
```
