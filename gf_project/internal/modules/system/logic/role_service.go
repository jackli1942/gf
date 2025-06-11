package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"

	"gf_project/internal/modules/system/model/do"
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"gf_project/internal/modules/system/model/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	// "github.com/gogf/gf/v2/util/gconv" // Not directly used in this version
)

type sSystemRole struct{}

// RoleService provides operations related to role management.
var RoleService = sSystemRole{}

// CreateRole creates a new role.
func (s *sSystemRole) CreateRole(ctx context.Context, in *input.RoleCreateInp) (roleId int64, err error) {
	count, err := dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Code, in.Code).Count()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to check role code existence")
	}
	if count > 0 {
		return 0, gerror.Newf("role code '%s' already exists", in.Code)
	}
	// Check for name uniqueness as well
	countName, errName := dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Name, in.Name).Count()
	if errName != nil {
		return 0, gerror.Wrap(errName, "failed to check role name existence")
	}
	if countName > 0 {
		return 0, gerror.Newf("role name '%s' already exists", in.Name)
	}

	newRole := do.SystemRole{
		Name:      in.Name,
		Code:      in.Code,
		Status:    in.Status,
		SortOrder: in.SortOrder,
		Remark:    in.Remark,
		CreatedAt: gtime.Now(),
		UpdatedAt: gtime.Now(),
	}

	result, err := dao.SystemRole.Ctx(ctx).Data(newRole).Insert()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to create role")
	}
	newRoleId, err := result.LastInsertId()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to retrieve last insert ID for role")
	}
	return newRoleId, nil
}

// UpdateRole updates an existing role.
func (s *sSystemRole) UpdateRole(ctx context.Context, in *input.RoleUpdateInp) (err error) {
	role, err := s.GetRoleById(ctx, in.Id)
	if err != nil {
		return err
	} // GetRoleById already wraps not found error

	if in.Code != role.Code { // If role code is being changed
		count, err := dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Code, in.Code).AndNot(dao.SystemRole.Columns().Id, in.Id).Count()
		if err != nil {
			return gerror.Wrap(err, "failed to check role code uniqueness on update")
		}
		if count > 0 {
			return gerror.Newf("role code '%s' already exists", in.Code)
		}
	}
	if in.Name != role.Name { // If role name is being changed
		countName, errName := dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Name, in.Name).AndNot(dao.SystemRole.Columns().Id, in.Id).Count()
		if errName != nil {
			return gerror.Wrap(errName, "failed to check role name uniqueness on update")
		}
		if countName > 0 {
			return gerror.Newf("role name '%s' already exists", in.Name)
		}
	}

	updateData := do.SystemRole{
		Name:      in.Name,
		Code:      in.Code,
		Status:    in.Status,
		SortOrder: in.SortOrder,
		Remark:    in.Remark,
		UpdatedAt: gtime.Now(),
	}

	_, err = dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Id, in.Id).Data(updateData).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to update role ID %d", in.Id)
	}

	if in.Code != role.Code {
		g.Log().Warningf(ctx, "Role code changed for role ID %d from '%s' to '%s'. Casbin policies using this role code might need manual adjustment or a migration script if role codes are used as subjects/objects in 'p' rules or as roles in 'g' rules.", in.Id, role.Code, in.Code)
		// For Casbin: If role code changes, all p rules (roleCode, path, method) and g rules (userId, roleCode) for the old roleCode need to be updated.
		// This is a complex operation and typically role codes are immutable.
		// If using hailaz adapter, it might not have direct support for renaming roles within Casbin policies automatically.
		// Manual update:
		// e, _ := CasbinEnforcer()
		// e.UpdateRoleForUsers(oldRoleCode, newRoleCode) // hypothetical
		// e.UpdatePermissionsForRole(oldRoleCode, newRoleCode) // hypothetical
	}
	return nil
}

// DeleteRole deletes one or more roles.
func (s *sSystemRole) DeleteRole(ctx context.Context, ids []uint64) (err error) {
	if len(ids) == 0 {
		return gerror.New("no role IDs provided for deletion")
	}

	e, casbinErr := CasbinEnforcer()
	if casbinErr != nil {
		return gerror.Wrap(casbinErr, "failed to get Casbin enforcer for deleting role policies")
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		for _, id := range ids {
			role := &entity.SystemRole{}
			// Fetch role details (especially role.Code) before deleting
			err := dao.SystemRole.Ctx(ctx).TX(tx).Where(dao.SystemRole.Columns().Id, id).Scan(&role)
			if err != nil {
				if errors.Is(err, gdb.ErrNoRows) {
					g.Log().Warningf(ctx, "Role with ID %d not found, skipping deletion.", id)
					continue
				}
				return gerror.Wrapf(err, "failed to retrieve role ID %d for deletion", id)
			}

			// 1. Delete from system_role (hard delete as SystemRole DO has no DeletedAt)
			_, err = dao.SystemRole.Ctx(ctx).TX(tx).Where(dao.SystemRole.Columns().Id, id).Delete()
			if err != nil {
				return gerror.Wrapf(err, "failed to delete role ID %d from DB", id)
			}

			// 2. Delete from system_user_role
			_, err = dao.SystemUserRole.Ctx(ctx).TX(tx).Where(dao.SystemUserRole.Columns().RoleId, id).Delete()
			if err != nil {
				return gerror.Wrapf(err, "failed to delete user-role associations for role ID %d", id)
			}

			// 3. Delete from system_role_menu (if used)
			_, err = dao.SystemRoleMenu.Ctx(ctx).TX(tx).Where(dao.SystemRoleMenu.Columns().RoleId, id).Delete()
			if err != nil {
				return gerror.Wrapf(err, "failed to delete role-menu associations for role ID %d", id)
			}

			// 4. Remove Casbin policies associated with this role's code
			if role.Code != "" {
				// Remove permission policies: p, role.Code, obj, act
				if _, errPol := e.RemoveFilteredPolicy(0, role.Code); errPol != nil {
					g.Log().Warningf(ctx, "Failed to remove Casbin 'p' policies for role code '%s': %v", role.Code, errPol)
				}
				// Remove grouping policies where this role is the role: g, user_id, role.Code
				if _, errGrp := e.RemoveFilteredGroupingPolicy(1, role.Code); errGrp != nil {
					g.Log().Warningf(ctx, "Failed to remove Casbin 'g' policies for role code '%s': %v", role.Code, errGrp)
				}
			}
		}
		// The hailaz adapter might auto-save on Add/Remove. If not, SavePolicy would be needed.
		// if errSave := e.SavePolicy(); errSave != nil {
		//    return gerror.Wrap(errSave, "Failed to save Casbin policies after deleting roles")
		// }
		return nil
	})
}

// GetRoleById retrieves a role by its ID.
func (s *sSystemRole) GetRoleById(ctx context.Context, id uint64) (role *entity.SystemRole, err error) {
	err = dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Id, id).Scan(&role)
	if err != nil {
		if errors.Is(err, gdb.ErrNoRows) {
			return nil, gerror.Newf("role with ID %d not found", id)
		}
		return nil, gerror.Wrapf(err, "failed to retrieve role by ID %d", id)
	}
	return role, nil
}

// GetRoleList retrieves a paginated list of roles.
func (s *sSystemRole) GetRoleList(ctx context.Context, in *input.RoleListInp) (list []*entity.SystemRole, total int, err error) {
	m := dao.SystemRole.Ctx(ctx).OmitEmptyWhere()
	if in.Name != "" {
		m = m.WhereLike(dao.SystemRole.Columns().Name, "%"+in.Name+"%")
	}
	if in.Code != "" { // Exact match for code often makes more sense
		m = m.Where(dao.SystemRole.Columns().Code, in.Code)
	}
	if in.Status != nil {
		m = m.Where(dao.SystemRole.Columns().Status, *in.Status)
	}

	total, err = m.Count()
	if err != nil {
		return nil, 0, gerror.Wrap(err, "failed to count roles")
	}
	if total == 0 {
		return []*entity.SystemRole{}, 0, nil
	}

	// Handle pagination defaults if Page or PageSize is 0
	page := in.Page
	pageSize := in.PageSize
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	} // Default page size

	err = m.Page(page, pageSize).OrderAsc(dao.SystemRole.Columns().SortOrder).OrderDesc(dao.SystemRole.Columns().Id).Scan(&list)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "failed to retrieve role list")
	}
	return list, total, nil
}

// UpdateRolePermissions updates the permissions (Casbin 'p' policies) for a given role code.
func (s *sSystemRole) UpdateRolePermissions(ctx context.Context, in *input.RoleSetPermissionsInp) error {
	if in.RoleCode == "" {
		return gerror.New("role code cannot be empty for updating permissions")
	}

	e, err := CasbinEnforcer()
	if err != nil {
		return gerror.Wrap(err, "failed to get Casbin enforcer for updating role permissions")
	}

	g.Log().Infof(ctx, "Updating permissions for roleCode: %s. New permissions count: %d", in.RoleCode, len(in.Permissions))

	// 1. Remove all existing 'p' policies for this roleCode
	if _, err = e.RemoveFilteredPolicy(0, in.RoleCode); err != nil {
		return gerror.Wrapf(err, "failed to remove existing Casbin policies for role '%s'", in.RoleCode)
	}
	g.Log().Debugf(ctx, "Removed existing policies for roleCode: %s", in.RoleCode)

	// 2. Add new policies
	if len(in.Permissions) > 0 {
		var policies [][]string
		for _, perm := range in.Permissions {
			if perm.Path == "" || perm.Method == "" {
				g.Log().Warningf(ctx, "Skipping invalid permission for role '%s': Path or Method is empty (%+v)", in.RoleCode, perm)
				continue
			}
			policies = append(policies, []string{in.RoleCode, perm.Path, perm.Method})
		}
		if len(policies) > 0 {
			added, addErr := e.AddPolicies(policies) // AddPolicies is variadic string arguments
			if addErr != nil {
				return gerror.Wrapf(addErr, "failed to add new Casbin policies for role '%s'", in.RoleCode)
			}
			if !added {
				g.Log().Warningf(ctx, "Casbin AddPolicies reported no policies were added for role '%s', though no error occurred (possibly all duplicates of existing or adapter issue).", in.RoleCode)
			} else {
				g.Log().Debugf(ctx, "Successfully added %d policies for roleCode: %s", len(policies), in.RoleCode)
			}
		}
	}

	// The hailaz adapter should auto-save. If not, e.SavePolicy() is needed.
	// if errSave := e.SavePolicy(); errSave != nil {
	//    return gerror.Wrap(errSave, "Failed to save Casbin policies after updating role permissions")
	// }
	return nil
}

// GetRolePermissions retrieves all 'p' policies for a given role code.
func (s *sSystemRole) GetRolePermissions(ctx context.Context, roleCode string) ([]input.PermissionRule, error) {
	if roleCode == "" {
		return nil, gerror.New("role code cannot be empty for retrieving permissions")
	}
	e, err := CasbinEnforcer()
	if err != nil {
		return nil, gerror.Wrap(err, "failed to get Casbin enforcer for retrieving role permissions")
	}

	casbinPolicies := e.GetFilteredPolicy(0, roleCode)

	permissions := make([]input.PermissionRule, 0, len(casbinPolicies))
	for _, p := range casbinPolicies {
		if len(p) >= 3 {
			permissions = append(permissions, input.PermissionRule{
				Path:   p[1],
				Method: p[2],
			})
		}
	}
	return permissions, nil
}

// UpdateRoleStatus updates a role's status.
func (s *sSystemRole) UpdateRoleStatus(ctx context.Context, id uint64, status uint) error {
	_, err := dao.SystemRole.Ctx(ctx).Where(dao.SystemRole.Columns().Id, id).Data(do.SystemRole{
		Status:    status,
		UpdatedAt: gtime.Now(),
	}).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to update status for role ID %d", id)
	}
	return nil
}
