package controller

import (
	"context"
	"gf_project/api/v1/system"                             // API Structs (Req/Res) for Menu
	"gf_project/internal/logic/middleware"                 // For ContextKeyUserId
	menuService "gf_project/internal/modules/system/logic" // Alias for menu service
	// commonLogic "gf_project/internal/logic" // Not directly used for responses now

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/util/gconv"
	// "github.com/gogf/gf/v2/frame/g"
)

type cSystemMenu struct{}

// MenuController is the exported instance of cSystemMenu.
var MenuController = cSystemMenu{}

// Create handles the creation of a new menu item.
// Corresponds to API struct system.MenuCreateReq
func (c *cSystemMenu) Create(ctx context.Context, req *system.MenuCreateReq) (res *system.MenuCreateRes, err error) {
	menuId, err := menuService.MenuService.CreateMenu(ctx, &req.MenuCreateInp)
	if err != nil {
		return nil, err
	}
	res = &system.MenuCreateRes{MenuId: menuId}
	return res, nil
}

// Update handles updating an existing menu item.
// Corresponds to API struct system.MenuUpdateReq
func (c *cSystemMenu) Update(ctx context.Context, req *system.MenuUpdateReq) (res *system.MenuUpdateRes, err error) {
	err = menuService.MenuService.UpdateMenu(ctx, &req.MenuUpdateInp)
	// res is nil for empty success response
	return nil, err
}

// Delete handles deleting a menu item.
// Corresponds to API struct system.MenuDeleteReq
func (c *cSystemMenu) Delete(ctx context.Context, req *system.MenuDeleteReq) (res *system.MenuDeleteRes, err error) {
	err = menuService.MenuService.DeleteMenu(ctx, req.Id)
	return nil, err
}

// List handles listing all menu items as a tree (for admin).
// Corresponds to API struct system.MenuListReq
func (c *cSystemMenu) List(ctx context.Context, req *system.MenuListReq) (res *system.MenuListRes, err error) {
	menuTree, err := menuService.MenuService.GetMenuAdminList(ctx, &req.MenuListInp)
	if err != nil {
		return nil, err
	}
	res = &system.MenuListRes{List: menuTree}
	return res, nil
}

// UserMenus handles retrieving the menu tree accessible by the current authenticated user.
// Corresponds to API struct system.UserMenuListReq
func (c *cSystemMenu) UserMenus(ctx context.Context, req *system.UserMenuListReq) (res *system.UserMenuListRes, err error) {
	userIdFromCtx := ctx.Value(middleware.ContextKeyUserId)
	if userIdFromCtx == nil {
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "User not authenticated: Missing UserId in Ctx for UserMenus")
	}
	userId := gconv.Uint64(userIdFromCtx)
	if userId == 0 { // Assuming 0 is not a valid/authenticated user ID
		return nil, gerror.NewCode(gcode.CodeNotAuthorized, "Invalid user ID in token for UserMenus (userId is 0)")
	}

	menuTree, err := menuService.MenuService.GetUserMenus(ctx, userId)
	if err != nil {
		return nil, err
	}
	res = &system.UserMenuListRes{List: menuTree}
	return res, nil
}

// Get handles retrieving a single menu item's details (for admin).
// Corresponds to API struct system.MenuGetReq
func (c *cSystemMenu) Get(ctx context.Context, req *system.MenuGetReq) (res *system.MenuGetRes, err error) {
	menu, err := menuService.MenuService.GetMenuById(ctx, req.Id)
	if err != nil {
		return nil, err
	}
	res = &system.MenuGetRes{SystemMenu: menu}
	return res, nil
}
