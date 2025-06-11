package system

import (
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"github.com/gogf/gf/v2/frame/g"
)

// Note: CommonPageRes is defined in user.go (api/v1/system/user.go)
// We will use that for responses. Ensure user.go is processed by codegen if using this.
// For clarity in this file, RoleListRes will explicitly list fields,
// but ideally, it would embed a shared CommonPageRes.

// RoleCreateReq is the API request structure for creating a role.
type RoleCreateReq struct {
	g.Meta `path:"/role/create" method:"post" tags:"System/RoleAdmin" summary:"Create a new role" security:"BearerAuth" group:"SystemRoleAdmin"`
	input.RoleCreateInp
}
type RoleCreateRes struct {
	RoleId uint64 `json:"roleId" dc:"Newly created Role ID"`
}

// RoleUpdateReq for updating a role by admin.
type RoleUpdateReq struct {
	g.Meta `path:"/role/update" method:"put" tags:"System/RoleAdmin" summary:"Update role details (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	input.RoleUpdateInp
}

// type RoleUpdateRes struct {} // Empty on success

// RoleDeleteReq for deleting role(s) by admin.
type RoleDeleteReq struct {
	g.Meta              `path:"/role/delete" method:"delete" tags:"System/RoleAdmin" summary:"Delete role(s) (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	input.RoleDeleteInp // Embeds Ids []uint64
}

// type RoleDeleteRes struct {} // Empty on success

// RoleListReq for listing roles (admin).
type RoleListReq struct {
	g.Meta `path:"/role/list" method:"get" tags:"System/RoleAdmin" summary:"List roles (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	input.RoleListInp
}

// RoleListRes defines the response structure for listing roles.
// Re-defining pagination fields here for clarity, or use CommonPageRes from user.go
type RoleListRes struct {
	List     []*entity.SystemRole `json:"list" dc:"List of roles"`
	Total    int                  `json:"total" dc:"Total number of roles"`
	Page     int                  `json:"page" dc:"Current page number"`
	PageSize int                  `json:"pageSize" dc:"Number of items per page"`
}

// RoleGetReq for getting a single role's details (admin).
type RoleGetReq struct {
	g.Meta `path:"/role/{id}" method:"get" tags:"System/RoleAdmin" summary:"Get role details by ID (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#Role ID must be a positive integer" dc:"Role ID"`
}

// RoleGetRes defines the response structure for getting a single role.
type RoleGetRes struct {
	*entity.SystemRole
}

// RoleSetPermissionsReq for admin setting permissions for a role.
type RoleSetPermissionsReq struct {
	g.Meta `path:"/role/set-permissions" method:"put" tags:"System/RoleAdmin" summary:"Set permissions for a role (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	input.RoleSetPermissionsInp
}

// type RoleSetPermissionsRes struct {} // Empty on success

// RoleGetPermissionsReq for admin getting permissions of a role.
type RoleGetPermissionsReq struct {
	g.Meta   `path:"/role/get-permissions/{roleCode}" method:"get" tags:"System/RoleAdmin" summary:"Get permissions for a role by code (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	RoleCode string `in:"path" v:"required#Role Code is required" dc:"Role Code"`
}

// RoleGetPermissionsRes defines the response structure for getting role permissions.
type RoleGetPermissionsRes struct {
	Permissions []input.PermissionRule `json:"permissions" dc:"List of permissions (path and method)"`
}

// RoleUpdateStatusReq for admin updating role status.
type RoleUpdateStatusReq struct {
	g.Meta `path:"/role/update-status" method:"put" tags:"System/RoleAdmin" summary:"Update role status (admin)" security:"BearerAuth" group:"SystemRoleAdmin"`
	input.RoleStatusInp
}

// type RoleUpdateStatusRes struct {} // Empty on success
