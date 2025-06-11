package input

// RoleCreateInp is the input parameter for creating a new role.
type RoleCreateInp struct {
	Name      string `json:"name"      v:"required|length:1,60#Role name is required|Role name length must be between 1 and 60 characters" dc:"Role Name"`
	Code      string `json:"code"      v:"required|length:1,60#Role code is required|Role code length must be between 1 and 60 characters" dc:"Role Code (e.g., admin, user_editor)"`
	Status    uint   `json:"status"    v:"in:0,1#Status must be 0 (active) or 1 (disabled)" dc:"Status (0:active, 1:disabled)"`
	SortOrder int    `json:"sortOrder" dc:"Sort Order for display"`
	Remark    string `json:"remark,omitempty"    v:"max-length:255#Remark cannot exceed 255 characters" dc:"Optional remark"`
	// DataScope uint   `json:"dataScope,omitempty" v:"in:1,2,3,4,5#Invalid data scope" dc:"Data Scope (1:All, 2:Custom, 3:Dept, 4:Dept & Sub, 5:Self) - Not fully implemented yet"`
}

// RoleUpdateInp is the input parameter for updating an existing role.
type RoleUpdateInp struct {
	Id        uint64 `json:"id"        v:"required|min:1#Role ID is required and must be a positive integer" dc:"Role ID"`
	Name      string `json:"name"      v:"required|length:1,60#Role name is required" dc:"Role Name"`
	Code      string `json:"code"      v:"required|length:1,60#Role code is required" dc:"Role Code"`
	Status    uint   `json:"status"    v:"in:0,1#Status must be 0 (active) or 1 (disabled)" dc:"Status"`
	SortOrder int    `json:"sortOrder" dc:"Sort Order"`
	Remark    string `json:"remark,omitempty"    v:"max-length:255#Remark cannot exceed 255 characters" dc:"Remark"`
	// DataScope uint   `json:"dataScope,omitempty"  v:"in:1,2,3,4,5#Invalid data scope" dc:"Data Scope - Not fully implemented yet"`
}

// RoleListInp is the input parameter for fetching a list of roles with optional filters and pagination.
type RoleListInp struct {
	Name     string `json:"name,omitempty"     dc:"Filter by role name (fuzzy match)"`
	Code     string `json:"code,omitempty"     dc:"Filter by role code (exact match)"`
	Status   *int   `json:"status,omitempty"   v:"in:0,1#Status must be 0 or 1 if provided" dc:"Filter by status (0:active, 1:disabled)"`
	Page     int    `json:"page,omitempty"     v:"min:0#Page number must be positive or zero for no pagination" dc:"Page number (for pagination, default 1 if 0)"`
	PageSize int    `json:"pageSize,omitempty" v:"min:0#Page size must be positive or zero for no pagination" dc:"Items per page (for pagination, default 10 if 0)"`
}

// PermissionRule defines a single permission (policy rule component).
type PermissionRule struct {
	Path   string `json:"path"   v:"required#Permission path (object) is required" dc:"Resource Path (e.g., /api/v1/system/user/list)"`
	Method string `json:"method" v:"required#Permission method (action) is required" dc:"HTTP Method (e.g., GET, POST, PUT, DELETE)"`
}

// RoleSetPermissionsInp is for setting permissions (Casbin policies) for a role.
// This will typically replace all existing permissions for the given role (by role code).
type RoleSetPermissionsInp struct {
	RoleCode    string           `json:"roleCode"    v:"required#Role Code is required" dc:"Role Code for which to set permissions"`
	Permissions []PermissionRule `json:"permissions" v:"array#Permissions must be an array" dc:"List of permissions to assign to the role"`
}

// RoleStatusInp is for changing a role's status.
type RoleStatusInp struct {
	Id     uint64 `json:"id"     v:"required|min:1#Role ID is required" dc:"Role ID"`
	Status uint   `json:"status" v:"required|in:0,1#Status is required and must be 0 or 1" dc:"Status (0:active, 1:disabled)"`
}

// RoleDeleteInp is for deleting one or more roles by ID.
type RoleDeleteInp struct {
	Ids []uint64 `json:"ids" v:"required|min-length:1#At least one Role ID is required" dc:"Array of Role IDs to delete"`
}
