package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"sort"    // For sorting menus by SortOrder
	"strconv" // For Casbin subject (userId)

	"gf_project/api/v1/system" // For MenuTreeItem struct
	"gf_project/internal/modules/system/model/do"
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"gf_project/internal/modules/system/model/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

type sSystemMenu struct{}

var MenuService = sSystemMenu{}

func buildMenuTree(ctx context.Context, parentId uint64, allMenus []*entity.SystemMenu) []*system.MenuTreeItem {
	tree := make([]*system.MenuTreeItem, 0)
	for _, menuEntity := range allMenus {
		if menuEntity.ParentId == parentId {
			// Create a mutable copy for children assignment if needed, or directly use.
			// For this, we create new MenuTreeItem and copy entity data.
			node := &system.MenuTreeItem{
				SystemMenu: &entity.SystemMenu{}, // Initialize embedded struct
			}
			_ = gconv.Struct(menuEntity, node.SystemMenu) // Copy fields

			childChildren := buildMenuTree(ctx, menuEntity.Id, allMenus)
			if len(childChildren) > 0 {
				node.Children = childChildren
			}
			tree = append(tree, node)
		}
	}
	sort.SliceStable(tree, func(i, j int) bool {
		if tree[i].Sort != tree[j].Sort {
			return tree[i].Sort < tree[j].Sort
		}
		return tree[i].Id < tree[j].Id // Fallback sort by ID for stable order
	})
	return tree
}

func (s *sSystemMenu) CreateMenu(ctx context.Context, in *input.MenuCreateInp) (menuId uint64, err error) {
	count, err := dao.SystemMenu.Ctx(ctx).Where(dao.SystemMenu.Columns().Code, in.Code).Count()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to check menu code existence")
	}
	if count > 0 {
		return 0, gerror.Newf("menu code (permission identifier) '%s' already exists", in.Code)
	}

	// Optional: Check for duplicate menu name under the same parent
	countName, errName := dao.SystemMenu.Ctx(ctx).Where(dao.SystemMenu.Columns().ParentId, in.ParentId).And(dao.SystemMenu.Columns().Name, in.Name).Count()
	if errName != nil {
		return 0, gerror.Wrap(errName, "failed to check menu name existence under parent")
	}
	if countName > 0 {
		return 0, gerror.Newf("menu name '%s' already exists under parent ID %d", in.Name, in.ParentId)
	}

	newMenu := do.SystemMenu{
		ParentId: in.ParentId, Name: in.Name, Code: in.Code, Icon: in.Icon,
		Route: in.Route, Component: in.Component, Redirect: in.Redirect,
		IsHidden: in.IsHidden, Type: in.Type, Status: in.Status, Sort: in.Sort,
		Remark: in.Remark, CreatedAt: gtime.Now(), UpdatedAt: gtime.Now(),
	}

	result, err := dao.SystemMenu.Ctx(ctx).Data(newMenu).Insert()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to create menu item")
	}

	newMenuId64, err := result.LastInsertId()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to retrieve last insert ID for menu")
	}
	return uint64(newMenuId64), nil
}

func (s *sSystemMenu) UpdateMenu(ctx context.Context, in *input.MenuUpdateInp) (err error) {
	menu, err := s.GetMenuById(ctx, in.Id)
	if err != nil {
		return err
	}

	if in.Code != menu.Code {
		count, errCode := dao.SystemMenu.Ctx(ctx).Where(dao.SystemMenu.Columns().Code, in.Code).AndNot(dao.SystemMenu.Columns().Id, in.Id).Count()
		if errCode != nil {
			return gerror.Wrap(errCode, "failed to check menu code uniqueness on update")
		}
		if count > 0 {
			return gerror.Newf("menu code '%s' already exists", in.Code)
		}
	}
	if in.Name != menu.Name && in.ParentId == menu.ParentId { // Only check name uniqueness if parent is the same
		countName, errName := dao.SystemMenu.Ctx(ctx).Where(dao.SystemMenu.Columns().ParentId, in.ParentId).And(dao.SystemMenu.Columns().Name, in.Name).AndNot(dao.SystemMenu.Columns().Id, in.Id).Count()
		if errName != nil {
			return gerror.Wrap(errName, "failed to check menu name uniqueness under parent on update")
		}
		if countName > 0 {
			return gerror.Newf("menu name '%s' already exists under parent ID %d", in.Name, in.ParentId)
		}
	}

	updateData := do.SystemMenu{
		Id: in.Id, ParentId: in.ParentId, Name: in.Name, Code: in.Code, Icon: in.Icon,
		Route: in.Route, Component: in.Component, Redirect: in.Redirect, IsHidden: in.IsHidden,
		Type: in.Type, Status: in.Status, Sort: in.Sort, Remark: in.Remark,
		UpdatedAt: gtime.Now(),
	}
	_, err = dao.SystemMenu.Ctx(ctx).Data(updateData).Where(dao.SystemMenu.Columns().Id, in.Id).Update()
	if err != nil {
		return gerror.Wrapf(err, "failed to update menu item ID %d", in.Id)
	}

	if in.Code != menu.Code {
		g.Log().Warningf(ctx, "Menu code for ID %d changed from '%s' to '%s'. Associated Casbin 'p' rules (permissions) might need manual update if menu codes are used as objects.", in.Id, menu.Code, in.Code)
	}
	return nil
}

func (s *sSystemMenu) getDescendantIdsAndCodes(ctx context.Context, parentId uint64, allMenus []*entity.SystemMenu) (ids []uint64, codes []string) {
	for _, menu := range allMenus {
		if menu.Id == parentId { // Add the parent itself first
			if !gconv.SliceContainsUint64(ids, menu.Id) {
				ids = append(ids, menu.Id)
				if menu.Code != "" {
					codes = append(codes, menu.Code)
				}
			}
		}
		if menu.ParentId == parentId {
			if !gconv.SliceContainsUint64(ids, menu.Id) { // Avoid processing same node multiple times if data is weird
				ids = append(ids, menu.Id)
				if menu.Code != "" {
					codes = append(codes, menu.Code)
				}
				childIds, childCodes := s.getDescendantIdsAndCodes(ctx, menu.Id, allMenus)
				ids = append(ids, childIds...)
				codes = append(codes, childCodes...)
			}
		}
	}
	return
}

func (s *sSystemMenu) DeleteMenu(ctx context.Context, id uint64) (err error) {
	allMenus, err := s.getAllMenus(ctx, nil) // Get all menus to build full hierarchy map
	if err != nil {
		return gerror.Wrap(err, "failed to get all menus for deletion pre-check")
	}

	idsToDelete, codesToDelete := s.getDescendantIdsAndCodes(ctx, id, allMenus)
	if !gconv.SliceContainsUint64(idsToDelete, id) { // Ensure the initial ID itself is included if it exists
		menu, _ := s.GetMenuById(ctx, id)
		if menu == nil {
			return gerror.Newf("menu with ID %d not found for deletion", id)
		}
		idsToDelete = append(idsToDelete, id)
		if menu.Code != "" {
			codesToDelete = append(codesToDelete, menu.Code)
		}
	}
	if len(idsToDelete) == 0 {
		return gerror.Newf("menu with ID %d not found or no descendants to delete", id)
	}

	e, casbinErr := CasbinEnforcer()
	if casbinErr != nil {
		return gerror.Wrap(casbinErr, "failed to get Casbin enforcer")
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		if len(idsToDelete) > 0 {
			_, err := dao.SystemMenu.Ctx(ctx).TX(tx).WhereIn(dao.SystemMenu.Columns().Id, idsToDelete).Delete()
			if err != nil {
				return gerror.Wrap(err, "failed to delete menu items from DB")
			}

			_, err = dao.SystemRoleMenu.Ctx(ctx).TX(tx).WhereIn(dao.SystemRoleMenu.Columns().MenuId, idsToDelete).Delete()
			if err != nil {
				g.Log().Warningf(ctx, "Could not clear system_role_menu for deleted menus: %v (non-fatal)", err)
			}
		}

		uniqueCodesToDelete := g.SliceStrUnique(codesToDelete) // Ensure unique codes for Casbin ops
		for _, menuCode := range uniqueCodesToDelete {
			if menuCode != "" {
				if _, errPol := e.RemoveFilteredPolicy(1, menuCode); errPol != nil { // fieldIndex 1 is obj (menu.Code)
					g.Log().Warningf(ctx, "Failed to remove Casbin 'p' policies where obj is '%s': %v", menuCode, errPol)
				}
			}
		}
		// Adapter might auto-save. If not: if errSave := e.SavePolicy(); errSave != nil { return gerror.Wrap(errSave, "failed to save Casbin policies"); }
		return nil
	})
}

func (s *sSystemMenu) GetMenuById(ctx context.Context, id uint64) (menu *entity.SystemMenu, err error) {
	err = dao.SystemMenu.Ctx(ctx).Where(dao.SystemMenu.Columns().Id, id).Scan(&menu)
	if err != nil {
		if errors.Is(err, gdb.ErrNoRows) {
			return nil, gerror.Newf("menu with ID %d not found", id)
		}
		return nil, gerror.Wrapf(err, "failed to retrieve menu by ID %d", id)
	}
	return menu, nil
}

func (s *sSystemMenu) getAllMenus(ctx context.Context, filter *input.MenuListInp) (list []*entity.SystemMenu, err error) {
	m := dao.SystemMenu.Ctx(ctx).OrderAsc(dao.SystemMenu.Columns().Sort).OrderAsc(dao.SystemMenu.Columns().Id)
	if filter != nil {
		if filter.Name != "" {
			m = m.WhereLike(dao.SystemMenu.Columns().Name, "%"+filter.Name+"%")
		}
		if filter.Status != nil {
			m = m.Where(dao.SystemMenu.Columns().Status, *filter.Status)
		}
	}
	err = m.Scan(&list)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to retrieve all menu items")
	}
	return list, nil
}

func (s *sSystemMenu) GetMenuAdminList(ctx context.Context, in *input.MenuListInp) ([]*system.MenuTreeItem, error) {
	allMenus, err := s.getAllMenus(ctx, in)
	if err != nil {
		return nil, err
	}
	return buildMenuTree(ctx, 0, allMenus), nil
}

func (s *sSystemMenu) GetUserMenus(ctx context.Context, userId uint64) ([]*system.MenuTreeItem, error) {
	// Get all active (status=0) menus first
	allActiveMenus, err := s.getAllMenus(ctx, &input.MenuListInp{Status: g.Uint(0)})
	if err != nil {
		return nil, gerror.Wrap(err, "failed to retrieve active menus for user menu generation")
	}

	e, casbinErr := CasbinEnforcer()
	if casbinErr != nil {
		return nil, gerror.Wrap(casbinErr, "failed to get Casbin enforcer for user menus")
	}

	userIdStr := strconv.FormatUint(userId, 10)
	permittedMenus := make([]*entity.SystemMenu, 0)

	for _, menu := range allActiveMenus {
		// Include Directories (M) and Menus (C) in the tree structure. Buttons/Functions (F) are for permission checks, not typically tree nodes.
		// Links (L, I) can be included if they are meant to be visible in navigation.
		if menu.Type == "M" || menu.Type == "C" || menu.Type == "L" || menu.Type == "I" {
			if menu.Code == "" { // Menus without a specific permission code might be considered generally accessible (if visible)
				g.Log().Debugf(ctx, "Menu '%s' (ID: %d) has no permission code, considered accessible by default if visible.", menu.Name, menu.Id)
				permittedMenus = append(permittedMenus, menu)
				continue
			}
			// Check Casbin: p, role_code, menu.Code, "access" (or specific action like "read", "view")
			// The enforcer will check if the user (via their roles) has this permission.
			allowed, enforceErr := e.Enforce(userIdStr, menu.Code, "access") // "access" action for menu visibility
			if enforceErr != nil {
				g.Log().Errorf(ctx, "Casbin enforce error for user %s, menu %s: %v", userIdStr, menu.Code, enforceErr)
				continue
			}
			if allowed {
				permittedMenus = append(permittedMenus, menu)
			}
		}
	}
	return buildMenuTree(ctx, 0, permittedMenus), nil
}
