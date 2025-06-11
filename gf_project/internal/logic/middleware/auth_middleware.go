package middleware

import (
	"context"
	"errors" // For jwt.NewValidationError
	"strings"
	"time" // For JWT expiry

	commonLogic "gf_project/internal/logic" // Renamed to avoid conflict with package name
	systemLogic "gf_project/internal/modules/system/logic"
	"gf_project/internal/modules/system/model/entity"

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/net/ghttp"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/text/gstr"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/golang-jwt/jwt/v4"
)

type sAuthMiddleware struct {
	Realm         string
	Key           []byte
	Timeout       time.Duration
	MaxRefresh    time.Duration
	IdentityKey   string // Key for user identity in JWT payload (e.g., "uid")
	TokenLookup   string
	TokenHeadName string
}

var Auth = sAuthMiddleware{}

const (
	ContextKeyUserObj    = "UserObject"
	ContextKeyUserId     = "UserId"
	ContextKeyUserRoles  = "UserRoles"
	ContextKeyUserClaims = "UserClaims"
)

type CustomClaims struct {
	UserID uint64   `json:"uid"`
	Roles  []string `json:"rol,omitempty"`
	jwt.RegisteredClaims
}

func (s *sAuthMiddleware) Init(ctx context.Context) error {
	jwtConfig := g.Cfg().MustGet(ctx, "jwt")
	if jwtConfig.IsNil() {
		return gerror.New("JWT configuration ('jwt' section) is missing in config.yaml")
	}
	s.Realm = jwtConfig.MustGet(ctx, "realm", "gf_project_realm").String()
	s.Key = []byte(jwtConfig.MustGet(ctx, "key", "").String())
	if len(s.Key) == 0 || string(s.Key) == "YourSecretKeyForJWT" || len(s.Key) < 32 {
		g.Log().Criticalf(ctx, "CRITICAL: JWT Key is empty, default, or too short (min 32 bytes recommended for HS256). Please set a strong 'jwt.key' in config.yaml!")
		// In a production system, you might want to return an error here to halt startup.
		// return gerror.New("Insecure JWT Key. Please update configuration with a strong key of at least 32 bytes.")
	}

	timeoutStr := jwtConfig.MustGet(ctx, "timeout", "2h").String()        // Default to 2 hours
	maxRefreshStr := jwtConfig.MustGet(ctx, "maxRefresh", "72h").String() // Default to 72 hours

	var err error
	s.Timeout, err = time.ParseDuration(timeoutStr)
	if err != nil {
		g.Log().Warningf(ctx, "Invalid JWT Timeout format: '%s', using 2 hours default. Error: %v", timeoutStr, err)
		s.Timeout = 2 * time.Hour
	}

	s.MaxRefresh, err = time.ParseDuration(maxRefreshStr)
	if err != nil {
		g.Log().Warningf(ctx, "Invalid JWT MaxRefresh format: '%s', using 72 hours default. Error: %v", maxRefreshStr, err)
		s.MaxRefresh = 72 * time.Hour
	}

	s.IdentityKey = "uid"
	s.TokenLookup = "header:Authorization"
	s.TokenHeadName = "Bearer"

	g.Log().Info(ctx, "JWT Authentication Middleware initialized.")
	return nil
}

func (s *sAuthMiddleware) Middleware(r *ghttp.Request) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		commonLogic.Respond(r, gcode.CodeNotAuthorized.Code(), "Authorization header is missing")
		return
	}

	parts := gstr.SplitAndTrim(authHeader, " ")
	if len(parts) != 2 || !gstr.Equal(parts[0], s.TokenHeadName) {
		commonLogic.Respond(r, gcode.CodeNotAuthorized.Code(), "Invalid Authorization header format (expected Bearer token)")
		return
	}
	tokenString := parts[1]

	claims := &CustomClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, gerror.Newf("unexpected signing method: %v", token.Header["alg"])
		}
		return s.Key, nil
	})

	if err != nil {
		errMsg := "Invalid or expired token"
		if ve, ok := err.(*jwt.ValidationError); ok {
			if ve.Errors&jwt.ValidationErrorMalformed != 0 {
				errMsg = "Malformed token"
			} else if ve.Errors&(jwt.ValidationErrorExpired|jwt.ValidationErrorNotValidYet) != 0 {
				errMsg = "Token is either expired or not active yet"
			} else if ve.Inner != nil { // More specific error
				errMsg = fmt.Sprintf("Token validation error: %s", ve.Inner.Error())
			}
		} else if !errors.Is(err, jwt.ErrTokenSignatureInvalid) { // Don't leak signature error details unless debugging
			g.Log().Debugf(r.Context(), "Unhandled JWT validation error: %+v", err)
		}
		commonLogic.Respond(r, gcode.CodeNotAuthorized.Code(), errMsg)
		return
	}

	if !token.Valid || claims.UserID == 0 {
		commonLogic.Respond(r, gcode.CodeNotAuthorized.Code(), "Invalid token claims")
		return
	}

	user, dbErr := systemLogic.UserService.GetUserById(r.Context(), claims.UserID)
	if dbErr != nil || user == nil {
		g.Log().Warningf(r.Context(), "User ID %d from valid token not found in DB: %v", claims.UserID, dbErr)
		commonLogic.Respond(r, gcode.CodeNotAuthorized.Code(), "User from token not found")
		return
	}
	if user.Status == 1 {
		commonLogic.Respond(r, gcode.CodeNotAuthorized.Code(), "User account is disabled")
		return
	}

	ctx := r.Context()
	ctx = context.WithValue(ctx, ContextKeyUserObj, user)
	ctx = context.WithValue(ctx, ContextKeyUserId, claims.UserID)
	ctx = context.WithValue(ctx, ContextKeyUserClaims, claims)
	if len(claims.Roles) > 0 {
		ctx = context.WithValue(ctx, ContextKeyUserRoles, claims.Roles)
	}
	r.SetCtx(ctx) // Update request context with new values

	r.Middleware.Next()
}

func (s *sAuthMiddleware) GenerateToken(ctx context.Context, userId uint64, userRoles []string) (tokenString string, expire gtime.Time, err error) {
	currentTime := time.Now()
	expireTime := currentTime.Add(s.Timeout)

	claims := CustomClaims{
		UserID: userId,
		Roles:  userRoles,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expireTime),
			IssuedAt:  jwt.NewNumericDate(currentTime),
			NotBefore: jwt.NewNumericDate(currentTime),
			Issuer:    s.Realm,
			Subject:   gconv.String(userId),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, signErr := token.SignedString(s.Key)
	if signErr != nil {
		return "", gtime.NewFromTime(expireTime), gerror.Wrap(signErr, "failed to sign JWT token")
	}
	return signedToken, gtime.NewFromTime(expireTime), nil
}

// --- Placeholder/Illustrative Handler Functions (not directly used by the custom Middleware above) ---

func AuthenticatorFromUserService(ctx context.Context, r *ghttp.Request) (interface{}, error) {
	var req input.UserLoginInp // Assuming UserLoginInp is defined in input package
	if err := r.Parse(&req); err != nil {
		return nil, gerror.WrapCode(gcode.CodeInvalidParameter, err, "Login parameter error")
	}
	user, err := systemLogic.UserService.UserLogin(ctx, &req)
	if err != nil {
		return nil, err
	}
	return user.Id, nil
}

func PayloadFuncHandler(data interface{}) g.Map {
	claims := g.Map{Auth.IdentityKey: data}
	// Example: Fetch roles and add to claims
	// userId := gconv.Uint64(data)
	// roles, _ := systemLogic.SomeRoleService.GetUserRoles(context.Background(), userId)
	// claims["roles"] = roles
	return claims
}

func IdentityHandlerFunc(ctx context.Context, r *ghttp.Request) interface{} {
	return r.Context().Value(ContextKeyUserId)
}

func LoginResponseHandler(ctx context.Context, r *ghttp.Request, code int, token string, expire time.Time) {
	commonLogic.Respond(r, gcode.CodeOK.Code(), "Login successful", g.Map{
		"token":  token,
		"expire": expire.Format(time.RFC3339),
	})
}

func RefreshResponseHandler(ctx context.Context, r *ghttp.Request, code int, token string, expire time.Time) {
	commonLogic.Respond(r, gcode.CodeOK.Code(), "Token refreshed", g.Map{
		"token":  token,
		"expire": expire.Format(time.RFC3339),
	})
}

func LogoutResponseHandler(ctx context.Context, r *ghttp.Request, code int) {
	commonLogic.Success(r, nil, "Logout successful")
}

func UnauthorizedHandler(ctx context.Context, r *ghttp.Request, code int, message string) {
	commonLogic.Respond(r, code, message)
}
