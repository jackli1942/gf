package controller

import (
	"context"
	"gf_project/api/v1/system"                             // API Structs (Req/Res)
	"gf_project/internal/logic/middleware"                 // For ContextKeyUserId
	userService "gf_project/internal/modules/system/logic" // Alias to user service

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/util/gconv"
)

type cSystemUser struct{}

var UserController = cSystemUser{}

// Create handles the creation of a new user.
func (c *cSystemUser) Create(ctx context.Context, req *system.UserCreateReq) (res *system.UserCreateRes, err error) {
	userId, err := userService.UserService.CreateUser(ctx, &req.UserCreateInp)
	if err != nil {
		return nil, err
	}
	res = &system.UserCreateRes{UserId: userId}
	return res, nil
}

// Login handles user login.
func (c *cSystemUser) Login(ctx context.Context, req *system.UserLoginReq) (res *system.UserLoginRes, err error) {
	userEntity, err := userService.UserService.UserLogin(ctx, &req.UserLoginInp)
	if err != nil {
		return nil, err
	}

	userRoleCodes, roleErr := userService.UserService.GetUserRoleCodes(ctx, userEntity.Id)
	if roleErr != nil {
		g.Log().Warningf(ctx, "Failed to fetch user roles for token (userId: %d): %v", userEntity.Id, roleErr)
		userRoleCodes = []string{}
	}

	tokenString, expireTime, tokenErr := middleware.Auth.GenerateToken(ctx, userEntity.Id, userRoleCodes)
	if tokenErr != nil {
		return nil, gerror.Wrap(tokenErr, "failed to generate token")
	}

	userEntity.PasswordHash = ""
	res = &system.UserLoginRes{
		User:  userEntity,
		Token: tokenString,
		// Expire: expireTime.Format(time.RFC3339), // Optionally include expiry in response
	}
	g.Log().Infof(ctx, "User '%s' (ID: %d) logged in successfully. Token expires at: %s", userEntity.Username, userEntity.Id, expireTime.String())
	return res, nil
}

// GetProfile handles retrieving the current logged-in user's profile.
func (c *cSystemUser) GetProfile(ctx context.Context, req *system.UserProfileReq) (res *system.UserProfileRes, err error) {
	userIdFromCtx := ctx.Value(middleware.ContextKeyUserId)
	if userIdFromCtx == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "User not authenticated: Missing UserId in Ctx")
	}
	userId := gconv.Uint64(userIdFromCtx)
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "Invalid user ID in token for GetProfile: UserId is 0")
	}

	user, err := userService.UserService.GetUserById(ctx, userId)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	res = &system.UserProfileRes{SystemUser: user}
	return res, nil
}

// UpdateProfile handles updating the current user's profile.
func (c *cSystemUser) UpdateProfile(ctx context.Context, req *system.UserUpdateProfileReq) (res *system.UserUpdateProfileRes, err error) {
	userIdFromCtx := ctx.Value(middleware.ContextKeyUserId)
	if userIdFromCtx == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "User not authenticated for UpdateProfile")
	}
	userId := gconv.Uint64(userIdFromCtx)
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "Invalid user ID in token for UpdateProfile")
	}

	err = userService.UserService.UpdateUserProfile(ctx, userId, &req.UserProfileUpdateInp)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ChangePassword handles a user changing their own password.
func (c *cSystemUser) ChangePassword(ctx context.Context, req *system.UserChangePasswordReq) (res *system.UserChangePasswordRes, err error) {
	userIdFromCtx := ctx.Value(middleware.ContextKeyUserId)
	if userIdFromCtx == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "User not authenticated for ChangePassword")
	}
	userId := gconv.Uint64(userIdFromCtx)
	if userId == 0 {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "Invalid user ID in token for ChangePassword")
	}

	err = userService.UserService.ChangeUserPassword(ctx, userId, req.OldPassword, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// List retrieves a list of users (admin).
func (c *cSystemUser) List(ctx context.Context, req *system.UserListReq) (res *system.UserListRes, err error) {
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	list, total, err := userService.UserService.GetUserList(ctx, &req.UserListInp)
	if err != nil {
		return nil, err
	}
	for _, user := range list {
		if user != nil {
			user.PasswordHash = ""
		}
	}
	res = &system.UserListRes{
		CommonPageRes: system.CommonPageRes{
			List:     list,
			Total:    total,
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	return res, nil
}

// Get handles retrieving a single user's details (admin).
func (c *cSystemUser) Get(ctx context.Context, req *system.UserGetReq) (res *system.UserGetRes, err error) {
	user, err := userService.UserService.GetUserById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	user.PasswordHash = ""
	res = &system.UserGetRes{SystemUser: user}
	return res, nil
}

// Update handles updating a user's details by an admin.
func (c *cSystemUser) Update(ctx context.Context, req *system.UserUpdateReq) (res *system.UserUpdateRes, err error) {
	err = userService.UserService.UpdateUser(ctx, &req.UserUpdateInp)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// Delete handles deleting a user by an admin.
func (c *cSystemUser) Delete(ctx context.Context, req *system.UserDeleteReq) (res *system.UserDeleteRes, err error) {
	err = userService.UserService.DeleteUser(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// ResetPassword handles an admin resetting a user's password.
func (c *cSystemUser) ResetPassword(ctx context.Context, req *system.AdminResetPasswordReq) (res *system.AdminResetPasswordRes, err error) {
	err = userService.UserService.AdminResetPassword(ctx, req.Id, req.NewPassword)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// UpdateStatus handles an admin changing a user's status.
func (c *cSystemUser) UpdateStatus(ctx context.Context, req *system.UserUpdateStatusReq) (res *system.UserUpdateStatusRes, err error) {
	err = userService.UserService.UpdateUserStatus(ctx, req.Id, req.Status)
	if err != nil {
		return nil, err
	}
	return nil, nil
}

// AssignRoles handles assigning roles to a user.
func (c *cSystemUser) AssignRoles(ctx context.Context, req *system.UserAssignRolesReq) (res *system.UserAssignRolesRes, err error) {
	err = userService.UserService.UpdateUserRoles(ctx, &req.UserAssignRoleInp)
	if err != nil {
		return nil, err
	}
	return nil, nil
}
