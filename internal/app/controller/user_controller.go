package controller

import (
	"context"
	"yuncms/internal/app/model"
	"yuncms/internal/app/service" // For User service
	// gfService "yuncms/internal/service" // For Token service if NewTokenService() was defined there

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/text/gstr"
	"yuncms/internal/middleware" // For GetCurrentUserIdFromCtx
)

type UserController struct{}

func NewUser() *UserController {
	return &UserController{}
}

// --- Create User ---
type UserCreateApiInput struct {
	Username string `json:"username" v:"required|length:3,30#user.usernameRequired|user.usernameLength" dc:"Username"`
	Password string `json:"password" v:"required|length:6,30#user.passwordRequired|user.passwordLength" dc:"Password"`
	Nickname string `json:"nickname" v:"required|length:1,30#user.nicknameRequired|user.nicknameLength" dc:"Nickname"`
	Email    string `json:"email,omitempty"    v:"email#user.emailFormat" dc:"Email address"`
	Phone    string `json:"phone,omitempty"    v:"phone-loose#user.phoneFormat" dc:"Phone number"`
	Avatar   string `json:"avatar,omitempty"   v:"url#user.avatarUrl" dc:"User avatar URL"`
	Status   *int   `json:"status,omitempty"   v:"in:0,1#user.statusInvalid" dc:"User status (0:disabled, 1:active, default:1)"`
	RoleIds  []uint `json:"roleIds,omitempty" dc:"List of role IDs to assign to the user"`
}

type UserCreateReq struct {
	g.Meta `path:"/users" method:"post" summary:"Create a new user" tags:"User Management"`
	UserCreateApiInput
}

type UserCreateRes struct {
	*model.User // Embed the user model, password will be excluded by json:"-" tag in model
}

func (c *UserController) Create(ctx context.Context, req *UserCreateReq) (res *UserCreateRes, err error) {
	serviceInput := service.UserCreateInput{
		Username: req.Username,
		Password: req.Password,
		Nickname: req.Nickname,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
		RoleIds:  req.RoleIds, // Pass RoleIds to service
	}
	createdUser, err := service.NewUserService().CreateUser(ctx, serviceInput)
	if err != nil {
		return nil, err
	}
	return &UserCreateRes{User: createdUser}, nil
}

// --- Get User By ID ---
type UserGetByIdReq struct {
	g.Meta `path:"/users/{id}" method:"get" summary:"Get user by ID" tags:"User Management"`
	Id     uint `json:"id" path:"id" v:"required|min:1#user.idRequired|user.idMin" dc:"User ID"`
}

type UserGetByIdRes struct {
	*model.User
}

func (c *UserController) GetById(ctx context.Context, req *UserGetByIdReq) (res *UserGetByIdRes, err error) {
	user, err := service.NewUserService().GetUserByID(ctx, req.Id)
	if err != nil {
		return nil, err // Service layer handles CodeNotFound with i18n
	}
	// Service GetUserByID now returns error if user is nil and not found.
	return &UserGetByIdRes{User: user}, nil
}

// --- List Users ---
type UserListReq struct {
	g.Meta `path:"/users" method:"get" summary:"List users" tags:"User Management"`
	Page   int `json:"page,omitempty"   in:"query" v:"min:1#pageMin" default:"1" dc:"Page number"`
	Size   int `json:"size,omitempty"   in:"query" v:"min:1|max:100#sizeMinMax" default:"10" dc:"Items per page"`
}

type UserListRes struct {
	*service.ListUsersOutput // Embed service output for flat JSON response
}

func (c *UserController) List(ctx context.Context, req *UserListReq) (res *UserListRes, err error) {
	serviceInput := service.ListUsersInput{
		Page: req.Page,
		Size: req.Size,
	}
	output, err := service.NewUserService().ListUsers(ctx, serviceInput)
	if err != nil {
		return nil, err
	}
	return &UserListRes{ListUsersOutput: output}, nil
}

// --- Update User ---
type UserUpdateApiInput struct { // Separate input DTO for body, to avoid embedding g.Meta in service.UserUpdateInput
	Nickname  *string `json:"nickname,omitempty" v:"length:1,30#user.nicknameLength" dc:"Nickname"`
	Password  *string `json:"password,omitempty" v:"length:6,30#user.passwordLength" dc:"New password (if changing)"`
	Email     *string `json:"email,omitempty"    v:"email#user.emailFormat" dc:"Email address"`
	Phone     *string `json:"phone,omitempty"    v:"phone-loose#user.phoneFormat" dc:"Phone number"`
	Avatar    *string `json:"avatar,omitempty"   v:"url#user.avatarUrl" dc:"User avatar URL"`
	Status    *int    `json:"status,omitempty"   v:"in:0,1#user.statusInvalid" dc:"User status (0:disabled, 1:active)"`
	RoleIds   *[]uint `json:"roleIds,omitempty" dc:"List of role IDs. If provided, replaces existing. If empty list, clears roles."`
}
type UserUpdateReq struct {
	g.Meta `path:"/users/{id}" method:"put" summary:"Update user by ID" tags:"User Management"`
	Id     uint `json:"id" path:"id" v:"required|min:1#user.idRequired|user.idMin" dc:"User ID"`
	UserUpdateApiInput
}

type UserUpdateRes struct {
	*model.User
}

func (c *UserController) Update(ctx context.Context, req *UserUpdateReq) (res *UserUpdateRes, err error) {
	serviceInput := service.UserUpdateInput{
		Nickname: req.Nickname,
		Password: req.Password,
		Email:    req.Email,
		Phone:    req.Phone,
		Avatar:   req.Avatar,
		Status:   req.Status,
		RoleIds:  req.RoleIds, // Pass RoleIds to service
	}

	updatedUserPartial, err := service.NewUserService().UpdateUser(ctx, req.Id, serviceInput) // service.UpdateUser returns *model.User
	if err != nil {
		errCode := gerror.Code(err)
		if errCode == gcode.CodeNotFound {
			return nil, gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "user.notFoundId", req.Id))
		} else if errCode == gcode.CodeBusinessValidationFailed {
			// Service layer should return errors with i18n keys or messages that are ready.
			// Example mapping if service error messages are too generic for API:
			errMsg := err.Error()
			if gstr.Contains(errMsg, g.I18n().T(ctx, "user.emailTakenBase")) { // Example key
				emailVal := ""
				if req.Email != nil { emailVal = *req.Email}
				return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, g.I18n().Tf(ctx, "user.emailTaken", emailVal))
			}
			return nil, err // Propagate service's business error
		}
		return nil, gerror.Wrapf(err, "Failed to update user with ID %d", req.Id)
	}
	return &UserUpdateRes{User: updatedUserPartial}, nil
}

// --- Delete User ---
type UserDeleteReq struct {
	g.Meta `path:"/users/{id}" method:"delete" summary:"Delete user by ID" tags:"User Management"`
	Id     uint `json:"id" path:"id" v:"required|min:1#user.idRequired|user.idMin" dc:"User ID"`
}

type UserDeleteRes struct {
	// Empty for 204-like response
}

func (c *UserController) Delete(ctx context.Context, req *UserDeleteReq) (res *UserDeleteRes, err error) {
	err = service.NewUserService().DeleteUser(ctx, req.Id)
	if err != nil {
		errCode := gerror.Code(err)
		// Service layer DeleteUser already returns CodeNotFound with i18n message
		// and CodeBusinessValidationFailed for hasChildren with i18n message.
		if errCode == gcode.CodeNotFound || errCode == gcode.CodeBusinessValidationFailed {
			return nil, err
		}
		return nil, gerror.Wrapf(err, "Failed to delete user with ID %d", req.Id)
	}
	return &UserDeleteRes{}, nil
}

// --- User Login ---
type UserLoginReq struct {
	g.Meta   `path:"/auth/login" method:"post" summary:"User login" tags:"Authentication"`
	Username string `json:"username" v:"required#auth.usernameRequired" dc:"Username"`
	Password string `json:"password" v:"required#auth.passwordRequired" dc:"Password"`
}

type UserLoginRes struct {
	Token    string      `json:"token" dc:"Authentication Token"`
	ExpireAt int64       `json:"expireAt" dc:"Token expiration timestamp"`
	User     *model.User `json:"user" dc:"User information"`
}

func (c *UserController) Login(ctx context.Context, req *UserLoginReq) (res *UserLoginRes, err error) {
	serviceInput := service.UserLoginInput{
		Username: req.Username,
		Password: req.Password,
	}
	loginOutput, err := service.NewUserService().LoginUser(ctx, serviceInput)
	if err != nil {
		return nil, err // Service layer handles i18n for auth errors
	}
	return &UserLoginRes{
		Token:    loginOutput.Token,
		ExpireAt: loginOutput.ExpireAt,
		User:     loginOutput.User,
	}, nil
}

// --- Admin Init Password ---
type UserInitPasswordReq struct {
	g.Meta   `path:"/users/{id}/init-password" method:"put" summary:"Admin initializes/resets user password" tags:"User Management"`
	Id       uint   `json:"id" path:"id" v:"required|min:1#user.idRequired|user.idMin" dc:"User ID"`
	Password string `json:"password,omitempty" v:"length:6,30#validation.passwordLength" dc:"New password. If empty, a default password from config (settings.defaultPassword) will be used."`
}

type UserInitPasswordRes struct {
	// Message string `json:"message,omitempty"` // Example: "Password reset successfully"
}

func (c *UserController) InitPassword(ctx context.Context, req *UserInitPasswordReq) (res *UserInitPasswordRes, err error) {
	// Validation of req.Id and req.Password (if provided) is handled by GoFrame based on struct tags.

	passwordToSet := req.Password
	if passwordToSet == "" {
		defaultPassword := g.Cfg().MustGet(ctx, "settings.defaultPassword", "DefaultPass123!").String() // Default from config
		passwordToSet = defaultPassword
		g.Log().Infof(ctx, "Admin is resetting password for UserID %d to default.", req.Id)
	}

	userService := service.NewUserService()
	err = userService.InitPassword(ctx, req.Id, passwordToSet)

	if err != nil {
		// Service layer InitPassword already returns gcode.CodeNotFound, gcode.CodeInvalidArgument,
		// or gcode.CodeValidationFailed with i18n-keyed messages.
		return nil, err // Propagate error directly
	}

	return &UserInitPasswordRes{}, nil
}

// --- User Modify Own Password ---
type UserModifyPasswordReq struct {
	g.Meta                  `path:"/users/modify-password" method:"put" summary:"User modifies their own password" tags:"User Management"`
	OldPassword             string `json:"oldPassword" v:"required#auth.oldPasswordRequired" dc:"Current password"`
	NewPassword             string `json:"newPassword" v:"required|length:6,30#validation.passwordLength" dc:"New password"`
	NewPasswordConfirmation string `json:"newPasswordConfirmation" v:"required|same:NewPassword#auth.newPasswordConfirmFailed" dc:"Confirm new password"`
}

type UserModifyPasswordRes struct {
	// Message string `json:"message,omitempty"` // Example: "Password changed successfully"
}

func (c *UserController) ModifyPassword(ctx context.Context, req *UserModifyPasswordReq) (res *UserModifyPasswordRes, err error) {
	// Validation of req fields is handled by GoFrame based on struct tags.

	userId := middleware.GetCurrentUserIdFromCtx(ctx) // Get user ID from JWT claims in context
	if userId == 0 {
		// This should ideally not happen if AuthMiddleware is effective and requires valid claims.
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, g.I18n().T(ctx, "auth.notLoggedIn")) // "auth.notLoggedIn = Please login first"
	}

	userService := service.NewUserService()
	err = userService.ModifyPassword(ctx, userId, req.OldPassword, req.NewPassword)

	if err != nil {
		// Service layer ModifyPassword already returns appropriate gcodes (CodeNotFound, CodeValidationFailed)
		// and uses i18n keys for messages (e.g., auth.oldPasswordIncorrect, validation.passwordLength).
		return nil, err // Propagate error directly
	}

	return &UserModifyPasswordRes{}, nil
}
```
