package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"os"
	"sort"
	"testing"

	"gf_project/api/v1/system" // For MenuTreeItem
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"gf_project/internal/modules/system/model/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gctx"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/test/gtest"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/gogf/gf/v2/util/grand"

	_ "github.com/gogf/gf/contrib/drivers/sqlite/v2"
	_ "github.com/hailaz/gf-casbin-adapter/v2"
)

var testMenuCtxGlobal = gctx.New()

func TestMain(m *testing.M) {
	ctx := testMenuCtxGlobal
	// Corrected project root calculation assuming tests are run from package dir 'logic'
	projectRoot := gfile.normalize(gfile.Pwd() + "/../../../")

	configDirPath := gfile.Join(projectRoot, "manifest/config")
	configFileActualPath := gfile.Join(configDirPath, "config.yaml")

	if gfile.Exists(configFileActualPath) {
		if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			adapter.SetPath(configDirPath)
			g.Log().Debugf(ctx, "TestMain (menu_service_test): Set config path to: %s", configDirPath)
		} else {
			g.Log().Warningf(ctx, "TestMain (menu_service_test): Config adapter is not *gcfg.AdapterFile, path not set.")
		}
	} else {
		g.Log().Warningf(ctx, "TestMain (menu_service_test): config.yaml not found at: %s", configFileActualPath)
	}

	modelPathFromConfig := g.Cfg().MustGet(ctx, "casbin.modelPath", "manifest/config/casbin_model.conf").String()
	resolvedModelPath := modelPathFromConfig
	if !gfile.IsAbs(modelPathFromConfig) {
		cfgAdapterPath := ""
		if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			cfgAdapterPath = adapterFile.GetPath()
		}
		if cfgAdapterPath != "" && gfile.Exists(gfile.Join(cfgAdapterPath, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(cfgAdapterPath, modelPathFromConfig)
		} else if projectRoot != "" && gfile.Exists(gfile.Join(projectRoot, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(projectRoot, modelPathFromConfig)
		} else if gfile.Exists(modelPathFromConfig) { // Check PWD if path is relative and not found via config/project root
			resolvedModelPath = modelPathFromConfig
		}
	}
	if !gfile.Exists(resolvedModelPath) {
		g.Log().Fatalf(ctx, "TestMain (menu_service_test): Casbin model file '%s' not found. Tests cannot proceed.", resolvedModelPath)
	}
	g.Cfg().Set("casbin.modelPath", resolvedModelPath) // Override for test session
	g.Log().Infof(ctx, "TestMain (menu_service_test): Effective Casbin model path: %s", resolvedModelPath)

	// Initialize Casbin Enforcer from the current 'logic' package
	if _, err := CasbinEnforcer(); err != nil {
		g.Log().Fatalf(testMenuCtxGlobal, "TestMain (menu_service_test): Failed to initialize Casbin enforcer: %v", err)
	}

	clearAllTestTablesForMenuTests(ctx) // Clear once before all tests in this file

	exitCode := m.Run()
	os.Exit(exitCode)
}

// clearAllTestTablesForMenuTests clears all tables that might be affected by menu tests.
func clearAllTestTablesForMenuTests(ctx context.Context) {
	tablesToClear := []string{
		dao.SystemUserRole.Table(),
		dao.SystemRoleMenu.Table(),
		dao.SystemMenu.Table(),
		dao.SystemRole.Table(),
		dao.SystemUser.Table(),
		"casbin_rule",
	}
	db := g.DB()
	allTablesInDb, _ := db.Ctx(ctx).Tables()
	existingTablesMap := make(map[string]bool)
	for _, t := range allTablesInDb {
		existingTablesMap[t] = true
	}

	for _, table := range tablesToClear {
		if !existingTablesMap[table] {
			g.Log().Debugf(ctx, "Table %s not found, skipping clear for menu tests.", table)
			continue
		}
		_, err := db.Ctx(ctx).Delete(table)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to clear table %s for menu tests: %v", table, err)
		}
		if db.GetConfig().DriverName == "sqlite" {
			_, errSeq := db.Ctx(ctx).Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", table))
			if errSeq != nil {
				g.Log().Debugf(ctx, "Failed to reset sqlite_sequence for table %s (menu tests): %v", table, errSeq)
			}
		}
	}
	e, err := CasbinEnforcer()
	if err == nil && e != nil {
		e.ClearPolicy()
		if loadErr := e.LoadPolicy(); loadErr != nil {
			g.Log().Warningf(ctx, "Error reloading Casbin policy after table clear (menu tests): %v", loadErr)
		}
	}
}

// Helper to create a role for menu tests, ensures unique code/name to avoid conflicts with other tests
func createTestRoleForMenuTest(ctx context.Context, suffix string) (code string, id uint64) {
	code = "role_menu_test_" + suffix
	name := "Role Menu Test " + suffix
	res, err := dao.SystemRole.Ctx(ctx).Data(g.Map{
		dao.SystemRole.Columns().Code:   code,
		dao.SystemRole.Columns().Name:   name,
		dao.SystemRole.Columns().Status: 0,
	}).Insert()
	if err != nil {
		g.Log().Fatalf(ctx, "Failed to create test role %s for menu test: %v", code, err)
	}
	idVal, _ := res.LastInsertId()
	return code, uint64(idVal)
}

func TestMenuService_CreateMenu(t *testing.T) {
	ctx := testMenuCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearAllTestTablesForMenuTests(ctx)
		uniqueSuffix := grand.S(4)
		createInput := &input.MenuCreateInp{
			ParentId: 0, Name: "Test Menu " + uniqueSuffix, Code: "test_menu_" + uniqueSuffix,
			Type: "M", Status: 0, Sort: 10, IsHidden: 0, // Status 0:Normal, IsHidden 0:No
		}
		menuId, err := MenuService.CreateMenu(ctx, createInput)
		t.AssertNil(err)
		t.AssertGT(menuId, 0)
		menu, _ := MenuService.GetMenuById(ctx, menuId)
		t.AssertNE(menu, nil)
		t.AssertEQ(menu.Name, createInput.Name)
	})
}

func TestMenuService_CreateMenu_DuplicateCode(t *testing.T) {
	ctx := testMenuCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearAllTestTablesForMenuTests(ctx)
		menuCode := "duplicate_menu_code"
		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "Menu A", Code: menuCode, Type: "C", Status: 0, IsHidden: 0})
		_, err := MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "Menu B", Code: menuCode, Type: "C", Status: 0, IsHidden: 0})
		t.AssertNE(err, nil)
		t.Assert(gerror.IsCode(gerror.Code(err), gcode.New(1, "", nil)), "Error code for duplicate menu code (should be custom or specific)") // Example for custom error code check
		t.Assert(gerror.Contains(err, "already exists"))
	})
}

func TestMenuService_UpdateMenu(t *testing.T) {
	ctx := testMenuCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearAllTestTablesForMenuTests(ctx)
		uniqueSuffix := grand.S(4)
		menuId, _ := MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "Original Menu " + uniqueSuffix, Code: "orig_menu_" + uniqueSuffix, Type: "M", Status: 0, IsHidden: 0})

		updateInput := &input.MenuUpdateInp{
			Id: menuId, ParentId: 0, Name: "Updated Menu Name " + uniqueSuffix, Code: "updated_menu_code_" + uniqueSuffix,
			Type: "C", Status: 0, Sort: 5, IsHidden: 0, Component: "/system/updated", Route: "/updated",
		}
		err := MenuService.UpdateMenu(ctx, updateInput)
		t.AssertNil(err)
		updatedMenu, _ := MenuService.GetMenuById(ctx, menuId)
		t.AssertEQ(updatedMenu.Name, updateInput.Name)
		t.AssertEQ(updatedMenu.Code, updateInput.Code)
	})
}

func TestMenuService_DeleteMenu(t *testing.T) {
	ctx := testMenuCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearAllTestTablesForMenuTests(ctx)
		e, _ := CasbinEnforcer()

		parentCode := "menu_parent_del"
		parentId, _ := MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "Parent Menu Del", Code: parentCode, Type: "M", Status: 0, IsHidden: 0})
		_, _ = e.AddPolicy("test_role_for_menu_del", parentCode, "access")

		childCode := "menu_child_del"
		childId, _ := MenuService.CreateMenu(ctx, &input.MenuCreateInp{ParentId: parentId, Name: "Child Menu Del", Code: childCode, Type: "C", Status: 0, IsHidden: 0})
		_, _ = e.AddPolicy("test_role_for_menu_del", childCode, "access")

		// If adapter autoSave is off, uncomment:
		// if errSave := e.SavePolicy(); errSave != nil { t.Fatalf("Failed to save policy for test setup: %v", errSave)}

		err := MenuService.DeleteMenu(ctx, parentId)
		t.AssertNil(err)

		pMenu, _ := MenuService.GetMenuById(ctx, parentId)
		cMenu, _ := MenuService.GetMenuById(ctx, childId)
		t.AssertNil(pMenu, "Parent menu should be deleted")
		t.AssertNil(cMenu, "Child menu should be deleted")

		// Reload policies from DB for verification as DeleteMenu modifies them
		// if errLoad := e.LoadPolicy(); errLoad != nil {t.Fatalf("Failed to reload policy for verification: %v", errLoad)}

		t.Assert(!e.HasPolicy("test_role_for_menu_del", parentCode, "access"), "Casbin policy for parent menu should be removed")
		t.Assert(!e.HasPolicy("test_role_for_menu_del", childCode, "access"), "Casbin policy for child menu should be removed")
	})
}

func TestMenuService_GetMenuAdminList(t *testing.T) {
	ctx := testMenuCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearAllTestTablesForMenuTests(ctx)
		m1Id, _ := MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "M1", Code: "m1", Type: "M", Status: 0, Sort: 10, IsHidden: 0})
		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{ParentId: m1Id, Name: "C1.2", Code: "c12", Type: "C", Status: 0, Sort: 20, IsHidden: 0}) // Child 2, higher sort
		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{ParentId: m1Id, Name: "C1.1", Code: "c11", Type: "C", Status: 0, Sort: 10, IsHidden: 0}) // Child 1, lower sort
		m2Id, _ := MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "M2", Code: "m2", Type: "M", Status: 0, Sort: 20, IsHidden: 0})
		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{ParentId: m2Id, Name: "C2.1", Code: "c21", Type: "C", Status: 0, Sort: 10, IsHidden: 0})

		tree, err := MenuService.GetMenuAdminList(ctx, &input.MenuListInp{})
		t.AssertNil(err)
		t.AssertEQ(len(tree), 2)

		var m1Node *system.MenuTreeItem
		for _, node := range tree {
			if node.Code == "m1" {
				m1Node = node
			}
		}
		t.AssertNE(m1Node, nil, "M1 node should exist")
		t.AssertEQ(len(m1Node.Children), 2)
		t.AssertEQ(m1Node.Children[0].Code, "c11", "First child of M1 should be C1.1 (sort 10)")
		t.AssertEQ(m1Node.Children[1].Code, "c12", "Second child of M1 should be C1.2 (sort 20)")
	})
}

func TestMenuService_GetUserMenus(t *testing.T) {
	ctx := testMenuCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearAllTestTablesForMenuTests(ctx)
		e, _ := CasbinEnforcer()

		userSuffix := grand.S(4)
		roleSuffix := grand.S(4)

		userId, _ := UserService.CreateUser(ctx, &input.UserCreateInp{Username: "menu_user_" + userSuffix, Password: "pw", Nickname: "MenuUser", Status: 0})
		userIdStr := gconv.String(userId)
		roleCode, roleId := createTestRoleForMenuTest(ctx, roleSuffix) // Uses helper

		assignInput := &input.UserAssignRoleInp{UserId: userId, RoleIds: []uint64{roleId}}
		_ = UserService.UpdateUserRoles(ctx, assignInput)

		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "Allowed Menu", Code: "menu_allowed", Type: "C", Status: 0, IsHidden: 0})
		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "Denied Menu", Code: "menu_denied", Type: "C", Status: 0, IsHidden: 0})
		_, _ = MenuService.CreateMenu(ctx, &input.MenuCreateInp{Name: "No Code Menu", Type: "C", Status: 0, IsHidden: 0}) // No code

		_, _ = e.AddPolicy(roleCode, "menu_allowed", "access")
		// e.SavePolicy() // If needed

		userMenuTree, err := MenuService.GetUserMenus(ctx, userId)
		t.AssertNil(err)

		t.AssertEQ(len(userMenuTree), 2, "User should see 2 menus (Allowed Menu + No Code Menu based on current logic)")

		foundAllowed := false
		foundNoCode := false
		for _, item := range userMenuTree {
			if item.Code == "menu_allowed" {
				foundAllowed = true
			}
			if item.Name == "No Code Menu" {
				foundNoCode = true
			} // Current logic includes menus with no code
		}
		t.Assert(foundAllowed, "Allowed Menu should be in the user's menu tree")
		t.Assert(foundNoCode, "No Code Menu should be in the user's menu tree (as per current service logic)")
	})
}
