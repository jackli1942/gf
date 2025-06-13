package service

import (
	"context"
	"errors"
	"time"
	// "yuncms/internal/app/model" // Not strictly needed for claims if using simplified UserTokenClaims

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv" // Added for gconv.String
	"github.com/golang-jwt/jwt/v5"
	// "github.com/google/uuid" // For JWT ID if needed, commented out for now
	"fmt" // For Redis key formatting
	"github.com/gogf/gf/v2/crypto/gmd5"
	"github.com/gogf/gf/v2/database/gredis"
)

const (
	defaultAppId = "default_app" // Simplified AppId for now
)

// TokenRedisInfo holds metadata for the token stored in Redis.
type TokenRedisInfo struct {
	ExpireAt     int64 `json:"exp"`
	RefreshAt    int64 `json:"ra"` // Timestamp of last refresh/activity
	RefreshCount int64 `json:"rc"` // How many times token has been refreshed (optional)
}

type sTokenService struct{}

// Ensure sTokenService implements ITokenService
var _ ITokenService = (*sTokenService)(nil)

func NewTokenService() ITokenService { // Return interface type
	return &sTokenService{}
}

// jwtCustomClaims are our custom claims incorporating UserTokenClaims and standard claims.
type jwtCustomClaims struct {
	UserTokenClaims
	jwt.RegisteredClaims
}

func (s *sTokenService) GenerateUserToken(ctx context.Context, userId uint, username string) (tokenString string, expireAt int64, err error) {
	jwtSecret := g.Cfg().MustGet(ctx, "jwt.secret", "your-default-jwt-secret-please-change").String()
	jwtExpireSeconds := g.Cfg().MustGet(ctx, "jwt.expireSeconds", int64(60*60*24)).Int64() // Default 1 day, ensure type matches

	nowTime := gtime.Now()
	expireAt = nowTime.Unix() + jwtExpireSeconds

	claims := jwtCustomClaims{
		UserTokenClaims: UserTokenClaims{
			UserId:   userId,
			Username: username,
		},
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Unix(expireAt, 0)),
			IssuedAt:  jwt.NewNumericDate(nowTime.Time),
			NotBefore: jwt.NewNumericDate(nowTime.Time),
			Issuer:    "yuncms",                 // Optional: Issuer
			Subject:   gconv.String(userId),     // Optional: Subject
			// ID:        guid.S(), 				// Optional: JWT ID, requires github.com/google/uuid
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err = token.SignedString([]byte(jwtSecret))
	if err != nil {
		g.Log().Errorf(ctx, "GenerateUserToken: Failed to sign JWT token for UserID %d: %v", userId, err)
		return "", 0, gerror.Wrap(err, g.I18n().T(ctx, "auth.tokenGenerationFailed"))
	}

	// Store token metadata in Redis
	redisAppId := g.Cfg().MustGet(ctx, "jwt.redisAppId", defaultAppId).String()
	authKey := s.getAuthKey(tokenString)
	tokenRedisKey := s.getTokenRedisKey(redisAppId, authKey)
	bindRedisKey := s.getBindRedisKey(redisAppId, userId)

	tokenInfo := TokenRedisInfo{
		ExpireAt:     expireAt,
		RefreshAt:    nowTime.Unix(),
		RefreshCount: 0,
	}

	redis := g.Redis() // Get default redis instance
	if redis == nil {
		g.Log().Error(ctx, "GenerateUserToken: Redis client is nil, cannot store token metadata.")
		// Depending on policy, you might allow login without Redis persistence or return an error
		return tokenString, expireAt, gerror.New("Redis client not available, token not persisted")
	}

	// Multi-login check
	allowMultiLogin := g.Cfg().MustGet(ctx, "jwt.multiLogin", true).Bool()
	if !allowMultiLogin {
		// Delete other tokens for this user
		existingTokenAuthKeys, err := redis.SMembers(ctx, bindRedisKey)
		if err != nil {
			g.Log().Errorf(ctx, "GenerateUserToken: Failed to get existing tokens for user %d from Redis: %v", userId, err)
			// Non-fatal, proceed with new token generation but log the issue
		} else {
			for _, oldAuthKey := range existingTokenAuthKeys {
				oldTokenRedisKey := s.getTokenRedisKey(redisAppId, oldAuthKey)
				_, _ = redis.Del(ctx, oldTokenRedisKey) // Ignore error on individual old token deletion
			}
			_, _ = redis.Del(ctx, bindRedisKey) // Delete the old bind set
		}
	}

	// Store token info and add to user's token set
	_, err = redis.SetEX(ctx, tokenRedisKey, tokenInfo, jwtExpireSeconds)
	if err != nil {
		g.Log().Errorf(ctx, "GenerateUserToken: Failed to store token info in Redis for UserID %d: %v", userId, err)
		return "", 0, gerror.Wrap(err, "Failed to persist token metadata")
	}

	_, err = redis.SAdd(ctx, bindRedisKey, authKey)
	if err != nil {
		g.Log().Errorf(ctx, "GenerateUserToken: Failed to bind token to user in Redis for UserID %d: %v", userId, err)
		// Attempt to clean up the token info if binding fails
		_, _ = redis.Del(ctx, tokenRedisKey)
		return "", 0, gerror.Wrap(err, "Failed to bind token to user")
	}
	// Set expiration for the bind key as well
	_, _ = redis.Expire(ctx, bindRedisKey, jwtExpireSeconds)


	return tokenString, expireAt, nil
}


// Helper methods for Redis keys
func (s *sTokenService) getAuthKey(tokenString string) string {
	return gmd5.MustEncryptString("yuncms:auth:" + tokenString)
}

func (s *sTokenService) getTokenRedisKey(appId string, authKey string) string {
	return fmt.Sprintf("token:%s:%s", appId, authKey)
}

func (s *sTokenService) getBindRedisKey(appId string, userId uint) string {
	return fmt.Sprintf("token_bind:%s:%d", appId, userId)
}


func (s *sTokenService) ParseUserToken(ctx context.Context, tokenString string) (claims *UserTokenClaims, err error) {
	if tokenString == "" {
		return nil, gerror.NewCode(gcode.New(401, "TOKEN_EMPTY", "Token string is empty"), g.I18n().T(ctx, "auth.tokenEmpty"))
	}
	jwtSecret := g.Cfg().MustGet(ctx, "jwt.secret", "your-default-jwt-secret-please-change").String()

	token, err := jwt.ParseWithClaims(tokenString, &jwtCustomClaims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, gerror.Newf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(jwtSecret), nil
	})

	if err != nil {
		// Handle common JWT errors specifically
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, gerror.NewCode(gcode.New(401, "TOKEN_EXPIRED", "Token has expired"), g.I18n().T(ctx, "auth.tokenExpired"))
		}
		if errors.Is(err, jwt.ErrTokenNotValidYet) {
			return nil, gerror.NewCode(gcode.New(401, "TOKEN_NOT_VALID_YET", "Token not active yet"), g.I18n().T(ctx, "auth.tokenNotValidYet"))
		}
		// For other parsing errors (e.g. signature invalid, malformed token)
		g.Log().Warningf(ctx, "ParseUserToken: Invalid token encountered: %v", err)
		return nil, gerror.NewCode(gcode.New(401, "TOKEN_INVALID", "Invalid token"), g.I18n().T(ctx, "auth.tokenInvalid"))
	}

	if jwtClaims, ok := token.Claims.(*jwtCustomClaims); ok && token.Valid {
		return &jwtClaims.UserTokenClaims, nil
	}
	return nil, gerror.NewCode(gcode.New(401, "TOKEN_INVALID_CLAIMS", "Invalid token claims"), g.I18n().T(ctx, "auth.tokenInvalidClaims"))
}

// Implementations for Redis-dependent methods
func (s *sTokenService) ValidateTokenInRedis(ctx context.Context, userId uint, tokenString string) (isValid bool, err error) {
	if tokenString == "" { // Should be caught by ParseUserToken first, but good check
		return false, gerror.NewCode(gcode.New(401, "TOKEN_EMPTY_FOR_VALIDATION", nil), "Token string is empty for validation")
	}
	redis := g.Redis()
	if redis == nil {
		g.Log().Error(ctx, "ValidateTokenInRedis: Redis client is nil.")
		return false, gerror.New("Redis client not available for token validation")
	}

	redisAppId := g.Cfg().MustGet(ctx, "jwt.redisAppId", defaultAppId).String()
	authKey := s.getAuthKey(tokenString)
	tokenRedisKey := s.getTokenRedisKey(redisAppId, authKey)
	bindRedisKey := s.getBindRedisKey(redisAppId, userId)

	// Check if token metadata exists
	exists, err := redis.Exists(ctx, tokenRedisKey)
	if err != nil {
		g.Log().Errorf(ctx, "ValidateTokenInRedis: Error checking token existence in Redis for UserID %d: %v", userId, err)
		return false, gerror.Wrap(err, "Failed to validate token existence from Redis")
	}
	if exists == 0 { // exists returns 0 if key does not exist, 1 if it exists
		return false, gerror.NewCode(gcode.New(401, "TOKEN_NOT_IN_REDIS", nil), "Token not found in active session store")
	}

	// Check if token is bound to the user
	isMember, err := redis.SIsMember(ctx, bindRedisKey, authKey)
	if err != nil {
		g.Log().Errorf(ctx, "ValidateTokenInRedis: Error checking token binding in Redis for UserID %d: %v", userId, err)
		return false, gerror.Wrap(err, "Failed to validate token binding from Redis")
	}
	if !isMember {
		return false, gerror.NewCode(gcode.New(401, "TOKEN_BINDING_MISMATCH", nil), "Token not bound to user session")
	}

	// Optionally: could refresh TTL of tokenRedisKey and bindRedisKey here if activity implies session extension
	// jwtExpireSeconds := g.Cfg().MustGet(ctx, "jwt.expireSeconds", int64(60*60*24)).Int64()
	// _, _ = redis.Expire(ctx, tokenRedisKey, jwtExpireSeconds)
	// _, _ = redis.Expire(ctx, bindRedisKey, jwtExpireSeconds)


	return true, nil
}

func (s *sTokenService) LogoutToken(ctx context.Context, userId uint, tokenString string) (err error) {
	if tokenString == "" {
		return gerror.NewCode(gcode.New(400, "TOKEN_EMPTY_FOR_LOGOUT", nil), "Token string is empty for logout")
	}
	redis := g.Redis()
	if redis == nil {
		g.Log().Error(ctx, "LogoutToken: Redis client is nil.")
		return gerror.New("Redis client not available for token logout")
	}

	redisAppId := g.Cfg().MustGet(ctx, "jwt.redisAppId", defaultAppId).String()
	authKey := s.getAuthKey(tokenString)
	tokenRedisKey := s.getTokenRedisKey(redisAppId, authKey)
	bindRedisKey := s.getBindRedisKey(redisAppId, userId)

	// Delete token metadata
	_, err = redis.Del(ctx, tokenRedisKey)
	if err != nil {
		g.Log().Warningf(ctx, "LogoutToken: Error deleting token key '%s' from Redis for UserID %d: %v", tokenRedisKey, userId, err)
		// Continue to attempt to remove from set
	}

	// Remove token from user's active token set
	_, err = redis.SRem(ctx, bindRedisKey, authKey)
	if err != nil {
		g.Log().Warningf(ctx, "LogoutToken: Error removing authKey '%s' from bindKey '%s' for UserID %d: %v", authKey, bindRedisKey, userId, err)
		// Depending on policy, this might be considered a partial failure.
		// For now, log and return nil if token key deletion was also attempted.
	}

	// Optional: If bindKey set becomes empty, delete it (mostly for cleanup, not critical for security)
	// card, _ := redis.SCard(ctx, bindRedisKey)
	// if card == 0 {
	// 	_, _ = redis.Del(ctx, bindRedisKey)
	// }

	return nil // Assume success even if some Redis cleanup steps had minor issues (logged)
}

func (s *sTokenService) KickUserTokens(ctx context.Context, userId uint) (err error) {
	redis := g.Redis()
	if redis == nil {
		g.Log().Error(ctx, "KickUserTokens: Redis client is nil.")
		return gerror.New("Redis client not available for kicking user tokens")
	}

	redisAppId := g.Cfg().MustGet(ctx, "jwt.redisAppId", defaultAppId).String()
	bindRedisKey := s.getBindRedisKey(redisAppId, userId)

	tokenAuthKeys, err := redis.SMembers(ctx, bindRedisKey)
	if err != nil {
		g.Log().Errorf(ctx, "KickUserTokens: Failed to get token auth keys for UserID %d from Redis: %v", userId, err)
		return gerror.Wrap(err, "Failed to retrieve user's active tokens for kicking")
	}

	if len(tokenAuthKeys) > 0 {
		tokenRedisKeysToDelete := make([]string, 0, len(tokenAuthKeys))
		for _, authKey := range tokenAuthKeys {
			tokenRedisKeysToDelete = append(tokenRedisKeysToDelete, s.getTokenRedisKey(redisAppId, authKey))
		}
		_, err = redis.Del(ctx, tokenRedisKeysToDelete...)
		if err != nil {
			g.Log().Warningf(ctx, "KickUserTokens: Error deleting some token keys from Redis for UserID %d: %v", userId, err)
			// Continue to delete the bind key
		}
	}

	// Delete the user's token binding set
	_, err = redis.Del(ctx, bindRedisKey)
	if err != nil {
		g.Log().Warningf(ctx, "KickUserTokens: Error deleting bindKey '%s' for UserID %d: %v", bindRedisKey, userId, err)
	}

	return nil // Assume best effort, log warnings for partial failures
}
```
