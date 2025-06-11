package controller

import (
	"context"
	"gf_project/api/v1/system" // API Structs (Req/Res) for Role
	// commonLogic "gf_project/internal/logic" // Common response utility - not directly used here
	roleService "gf_project/internal/modules/system/logic" // Alias for role service
	// "github.com/gogf/gf/v2/frame/g" // Not directly needed if only using context and API structs
)

type cSystemRole struct{}

// RoleController is the exported instance of cSystemRole.
var RoleController = cSystemRole{}

// Create handles the creation of a new role.
// Corresponds to API struct system.RoleCreateReq
func (c *cSystemRole) Create(ctx context.Context, req *system.RoleCreateReq) (res *system.RoleCreateRes, err error) {
	roleId, err := roleService.RoleService.CreateRole(ctx, &req.RoleCreateInp)
	if err != nil {
		return nil, err
	}
	res = &system.RoleCreateRes{RoleId: roleId}
	return res, nil
}

// Update handles updating an existing role.
// Corresponds to API struct system.RoleUpdateReq
func (c *cSystemRole) Update(ctx context.Context, req *system.RoleUpdateReq) (res *system.RoleUpdateRes, err error) {
	err = roleService.RoleService.UpdateRole(ctx, &req.RoleUpdateInp)
	// res is nil for empty success response, GoFrame handles this.
	return nil, err
}

// Delete handles deleting one or more roles.
// Corresponds to API struct system.RoleDeleteReq
func (c *cSystemRole) Delete(ctx context.Context, req *system.RoleDeleteReq) (res *system.RoleDeleteRes, err error) {
	err = roleService.RoleService.DeleteRole(ctx, req.Ids) // req.Ids from embedded input.RoleDeleteInp
	return nil, err
}

// List retrieves a list of roles.
// Corresponds to API struct system.RoleListReq
func (c *cSystemRole) List(ctx context.Context, req *system.RoleListReq) (res *system.RoleListRes, err error) {
	// Default pagination if not provided in req (or rely on validation tags in input struct)
	if req.Page == 0 {
		req.Page = 1
	}
	if req.PageSize == 0 {
		req.PageSize = 10
	}

	list, total, err := roleService.RoleService.GetRoleList(ctx, &req.RoleListInp)
	if err != nil {
		return nil, err
	}
	res = &system.RoleListRes{ // Assuming RoleListRes has fields List, Total, Page, PageSize
		List:     list,
		Total:    total,
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	return res, nil
}

// Get handles retrieving a single role's details.
// Corresponds to API struct system.RoleGetReq
func (c *cSystemRole) Get(ctx context.Context, req *system.RoleGetReq) (res *system.RoleGetRes, err error) {
	role, err := roleService.RoleService.GetRoleById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &system.RoleGetRes{SystemRole: role}
	return res, nil
}

// UpdatePermissions handles setting permissions for a role.
// Corresponds to API struct system.RoleSetPermissionsReq
func (c *cSystemRole) UpdatePermissions(ctx context.Context, req *system.RoleSetPermissionsReq) (res *system.RoleSetPermissionsRes, err error) {
	err = roleService.RoleService.UpdateRolePermissions(ctx, &req.RoleSetPermissionsInp) // Pass embedded input
	return nil, err
}

// GetPermissions handles retrieving permissions for a role.
// Corresponds to API struct system.RoleGetPermissionsReq
func (c *cSystemRole) GetPermissions(ctx context.Context, req *system.RoleGetPermissionsReq) (res *system.RoleGetPermissionsRes, err error) {
	permissions, err := roleService.RoleService.GetRolePermissions(ctx, req.RoleCode)
	if err != nil {
		return nil, err
	}
	res = &system.RoleGetPermissionsRes{Permissions: permissions}
	return res, nil
}

// UpdateStatus handles updating a role's status.
// Corresponds to API struct system.RoleUpdateStatusReq
func (c *cSystemRole) UpdateStatus(ctx context.Context, req *system.RoleUpdateStatusReq) (res *system.RoleUpdateStatusRes, err error) {
	err = roleService.RoleService.UpdateRoleStatus(ctx, req.Id, req.Status)
	return nil, err
}
