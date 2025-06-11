package system

import (
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input" // Assuming input structs are here
	"github.com/gogf/gf/v2/frame/g"
)

// CommonPageRes is a standard structure for paginated list responses.
type CommonPageRes struct {
	List     interface{} `json:"list" dc:"List of items"`
	Total    int         `json:"total" dc:"Total number of items"`
	Page     int         `json:"page" dc:"Current page number"`
	PageSize int         `json:"pageSize" dc:"Number of items per page"`
}

// UserCreateReq defines the request structure for creating a user.
type UserCreateReq struct {
	g.Meta `path:"/user/create" method:"post" tags:"System/User" summary:"Create a new user" group:"SystemUser"`
	input.UserCreateInp
}

// UserCreateRes defines the response structure for creating a user.
type UserCreateRes struct {
	UserId int64 `json:"userId" dc:"Newly created User ID"`
}

// UserLoginReq defines the request structure for user login.
type UserLoginReq struct {
	g.Meta `path:"/user/login" method:"post" tags:"System/User" summary:"User login" group:"SystemUser"`
	input.UserLoginInp
}

// UserLoginRes defines the response structure for user login.
type UserLoginRes struct {
	User  *entity.SystemUser `json:"user" dc:"User details (excluding sensitive info)"` // Ensure password is not sent
	Token string             `json:"token" dc:"JWT authentication token"`
}

// UserProfileReq defines the request structure for getting current user's profile.
type UserProfileReq struct {
	g.Meta `path:"/user/profile" method:"get" tags:"System/User" summary:"Get current user profile" security:"BearerAuth" group:"SystemUser"`
	// No request body parameters needed, user identified by token
}

// UserProfileRes defines the response structure for getting current user's profile.
type UserProfileRes struct {
	*entity.SystemUser // Embed user entity directly, ensure password hash is cleared by service/controller
}

// UserUpdateProfileReq defines the request structure for updating current user's profile.
type UserUpdateProfileReq struct {
	g.Meta `path:"/user/profile" method:"put" tags:"System/User" summary:"Update current user profile" security:"BearerAuth" group:"SystemUser"`
	input.UserProfileUpdateInp
}

// UserUpdateProfileRes can be an empty success response or return the updated profile.
// For simplicity, often an empty success response is used, client assumes input values are set.
// type UserUpdateProfileRes struct {} // Or return UserProfileRes

// UserChangePasswordReq defines the request structure for user changing their own password.
type UserChangePasswordReq struct {
	g.Meta `path:"/user/change-password" method:"put" tags:"System/User" summary:"Change current user password" security:"BearerAuth" group:"SystemUser"`
	input.UserChangePasswordInp
}

// type UserChangePasswordRes struct {} // Empty on success

// UserListReq defines the request structure for listing users (admin).
type UserListReq struct {
	g.Meta `path:"/user/list" method:"get" tags:"System/UserAdmin" summary:"List users (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	input.UserListInp
}

// UserListRes defines the response structure for listing users.
type UserListRes struct {
	CommonPageRes
}

// UserGetReq defines the request structure for getting a single user's details (admin).
type UserGetReq struct {
	g.Meta `path:"/user/{id}" method:"get" tags:"System/UserAdmin" summary:"Get user details by ID (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#User ID must be a positive integer" dc:"User ID"`
}

// UserGetRes defines the response structure for getting a single user.
type UserGetRes struct {
	*entity.SystemUser // Embed user entity, ensure password hash is cleared
}

// UserUpdateReq defines the request structure for updating a user by admin.
type UserUpdateReq struct {
	g.Meta              `path:"/user/update" method:"put" tags:"System/UserAdmin" summary:"Update user details (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	input.UserUpdateInp // Includes Id in the body
}

// type UserUpdateRes struct {} // Empty on success

// UserDeleteReq defines the request structure for deleting a user by admin.
type UserDeleteReq struct {
	g.Meta `path:"/user/delete/{id}" method:"delete" tags:"System/UserAdmin" summary:"Delete user (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#User ID must be a positive integer" dc:"User ID"`
}

// type UserDeleteRes struct {} // Empty on success

// AdminResetPasswordReq defines the request for admin resetting user password.
type AdminResetPasswordReq struct {
	g.Meta `path:"/user/reset-password" method:"put" tags:"System/UserAdmin" summary:"Reset user password (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	input.AdminResetUserPasswordInp
}

// type AdminResetPasswordRes struct {} // Empty on success

// UserUpdateStatusReq defines the request for admin updating user status.
type UserUpdateStatusReq struct {
	g.Meta `path:"/user/update-status" method:"put" tags:"System/UserAdmin" summary:"Update user status (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	input.UserStatusInp
}

// type UserUpdateStatusRes struct {} // Empty on success

// UserAssignRolesReq defines the request for admin assigning roles to a user.
type UserAssignRolesReq struct {
	g.Meta `path:"/user/assign-roles" method:"put" tags:"System/UserAdmin" summary:"Assign roles to user (admin)" security:"BearerAuth" group:"SystemUserAdmin"`
	input.UserAssignRoleInp
}

// type UserAssignRolesRes struct {} // Empty on success
