package input

// UserCreateInp is the input parameter for creating a new user.
// Based on devinggo's sys_user table structure.
type UserCreateInp struct {
	Username string `json:"username" v:"required|length:2,60#Username is required|Username length must be between 2 and 60 characters" dc:"Username"`
	Password string `json:"password" v:"required|length:6,30#Password is required|Password length must be between 6 and 30 characters" dc:"Password (will be hashed)"`
	Nickname string `json:"nickname" v:"required|length:1,60#Nickname is required|Nickname length must be between 1 and 60 characters" dc:"Nickname"`
	Email    string `json:"email,omitempty"    v:"email#Invalid email format" dc:"Email address"`
	Mobile   string `json:"mobile,omitempty"   v:"phone#Invalid mobile phone number format" dc:"Mobile phone number"`
	DeptId   uint64 `json:"deptId,omitempty"   v:"min:0#Department ID must be a positive integer or zero" dc:"Department ID"`
	Avatar   string `json:"avatar,omitempty"   v:"url#Avatar must be a valid URL" dc:"User Avatar URL"`
	Gender   *int   `json:"gender,omitempty"   v:"in:0,1,2#Gender must be 0, 1, or 2" dc:"Gender (0:Unknown, 1:Male, 2:Female)"`
	Status   uint   `json:"status"   v:"in:0,1#Status must be 0 (active) or 1 (disabled)" dc:"Status (0:active, 1:disabled)"`
	Remark   string `json:"remark,omitempty"   v:"max-length:255#Remark cannot exceed 255 characters" dc:"Remark"`
}

// UserUpdateInp is the input parameter for updating an existing user.
// Username is typically not updatable. Password update is a separate action.
type UserUpdateInp struct {
	Id       uint64 `json:"id"       v:"required|min:1#User ID is required and must be a positive integer" dc:"User ID"`
	Nickname string `json:"nickname" v:"required|length:1,60#Nickname is required|Nickname length must be between 1 and 60 characters" dc:"Nickname"`
	Email    string `json:"email,omitempty"    v:"email#Invalid email format" dc:"Email address"`
	Mobile   string `json:"mobile,omitempty"   v:"phone#Invalid mobile phone number format" dc:"Mobile phone number"`
	DeptId   uint64 `json:"deptId,omitempty"   v:"min:0#Department ID must be a positive integer or zero" dc:"Department ID"`
	Avatar   string `json:"avatar,omitempty"   v:"url#Avatar must be a valid URL" dc:"User Avatar URL"`
	Gender   *int   `json:"gender,omitempty"   v:"in:0,1,2#Gender must be 0, 1, or 2" dc:"Gender (0:Unknown, 1:Male, 2:Female)"`
	Status   uint   `json:"status"   v:"in:0,1#Status must be 0 (active) or 1 (disabled)" dc:"Status (0:active, 1:disabled)"`
	Remark   string `json:"remark,omitempty"   v:"max-length:255#Remark cannot exceed 255 characters" dc:"Remark"`
}

// UserChangePasswordInp is the input parameter for a user changing their own password.
type UserChangePasswordInp struct {
	OldPassword string `json:"oldPassword" v:"required#Old password is required" dc:"Old Password"`
	NewPassword string `json:"newPassword" v:"required|length:6,30#New password must be between 6 and 30 characters" dc:"New Password"`
}

// AdminResetUserPasswordInp is the input parameter for an admin resetting a user's password.
type AdminResetUserPasswordInp struct {
	Id          uint64 `json:"id"          v:"required|min:1#User ID is required" dc:"User ID"`
	NewPassword string `json:"newPassword" v:"required|length:6,30#New password must be between 6 and 30 characters" dc:"New Password"`
}

// UserLoginInp is the input parameter for user login.
type UserLoginInp struct {
	Username string `json:"username" v:"required#Username is required" dc:"Username"`
	Password string `json:"password" v:"required#Password is required" dc:"Password"`
}

// UserListInp is the input parameter for fetching a list of users (with pagination and filters).
type UserListInp struct {
	Username string  `json:"username,omitempty" dc:"Username to filter by"`
	Mobile   string  `json:"mobile,omitempty"   dc:"Mobile number to filter by"`
	Status   *int    `json:"status,omitempty"   v:"in:0,1#Status must be 0 or 1 if provided" dc:"Status to filter by (0:active, 1:disabled)"`
	DeptId   *uint64 `json:"deptId,omitempty"   dc:"Department ID to filter by"`
	Page     int     `json:"page"     v:"required|min:1#Page number must be at least 1" dc:"Page number (for pagination)"`
	PageSize int     `json:"pageSize" v:"required|min:1|max:100#Page size must be between 1 and 100" dc:"Items per page (for pagination)"`
}

// UserProfileUpdateInp is for updating the current user's profile information.
// Based on devinggo's sys_user table structure.
type UserProfileUpdateInp struct {
	Nickname string `json:"nickname,omitempty" v:"length:1,60#Nickname length must be between 1 and 60 characters" dc:"Nickname"`
	Email    string `json:"email,omitempty"    v:"email#Invalid email format" dc:"Email"`
	Mobile   string `json:"mobile,omitempty"   v:"phone#Invalid mobile format" dc:"Mobile"`
	Gender   *int   `json:"gender,omitempty"   v:"in:0,1,2#Gender must be 0, 1, or 2" dc:"Gender (0:Unknown, 1:Male, 2:Female)"`
	Avatar   string `json:"avatar,omitempty"   v:"url#Invalid avatar URL" dc:"Avatar URL"`
	Signed   string `json:"signed,omitempty"   v:"max-length:255#Signature too long" dc:"Personal Signature"`
}

// UserStatusInp is for changing user status.
type UserStatusInp struct {
	Id     uint64 `json:"id" v:"required|min:1#User ID is required" dc:"User ID"`
	Status uint   `json:"status" v:"required|in:0,1#Status is required and must be 0 or 1" dc:"Status (0:active, 1:disabled)"`
}

// UserAssignRoleInp is for assigning roles to a user.
type UserAssignRoleInp struct {
	UserId  uint64   `json:"userId" v:"required|min:1#User ID is required" dc:"User ID"`
	RoleIds []uint64 `json:"roleIds" v:"array#Role IDs must be an array" dc:"Array of Role IDs"`
}

// UserGetProfileOut is the output parameter for getting user profile.
// This is an example, you might have a more specific output struct or use entity directly.
// For now, we might not need a specific output struct if entity.User is sufficient.
// type UserGetProfileOut entity.User // Example, assuming entity.User exists and is suitable

// UserAssignRoleInp is for assigning roles to a user.
type UserAssignRoleInp struct {
	UserId  uint64   `json:"userId"  v:"required|min:1#User ID is required" dc:"User ID"`
	RoleIds []uint64 `json:"roleIds" v:"array#Role IDs must be an array" dc:"Array of Role IDs to assign (can be empty to remove all roles)"`
}
