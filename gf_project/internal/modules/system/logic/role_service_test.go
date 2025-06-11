package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"os"
	"testing"
	// "time"

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

var testRoleCtxGlobal = gctx.New()

func TestMain(m *testing.M) {
	ctx := testRoleCtxGlobal
	projectRoot := gfile.normalize(gfile.Pwd() + "/../../../")

	configDirPath := gfile.Join(projectRoot, "manifest/config")
	configFileActualPath := gfile.Join(configDirPath, "config.yaml")

	if gfile.Exists(configFileActualPath) {
		if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			adapter.SetPath(configDirPath)
			g.Log().Debugf(ctx, "TestMain (role_service_test): Set config path to: %s", configDirPath)
		} else {
			g.Log().Warningf(ctx, "TestMain (role_service_test): Config adapter is not *gcfg.AdapterFile, path not set.")
		}
	} else {
		g.Log().Warningf(ctx, "TestMain (role_service_test): config.yaml not found at: %s", configFileActualPath)
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
		} else if gfile.Exists(modelPathFromConfig) {
			resolvedModelPath = modelPathFromConfig
		}
	}
	if !gfile.Exists(resolvedModelPath) {
		g.Log().Fatalf(ctx, "TestMain (role_service_test): Casbin model file '%s' not found. Tests cannot proceed.", resolvedModelPath)
	}
	g.Cfg().Set("casbin.modelPath", resolvedModelPath)
	g.Log().Infof(ctx, "TestMain (role_service_test): Effective Casbin model path: %s", resolvedModelPath)

	if _, err := CasbinEnforcer(); err != nil { // From current package 'logic'
		g.Log().Fatalf(ctx, "TestMain (role_service_test): Failed to initialize Casbin enforcer: %v", err)
	}

	clearTestRoleAndRelatedTables(ctx) // Clear once before all tests in this file

	exitCode := m.Run()
	os.Exit(exitCode)
}

func clearTestRoleAndRelatedTables(ctx context.Context) {
	tablesToClear := []string{
		dao.SystemUserRole.Table(),
		dao.SystemRoleMenu.Table(), // Assuming this join table exists
		dao.SystemRole.Table(),
		dao.SystemUser.Table(), // Clear users also as DeleteRole test creates one
		"casbin_rule",
	}
	db := g.DB()
	allTables, _ := db.Ctx(ctx).Tables()
	existingTablesMap := make(map[string]bool)
	for _, t := range allTables {
		existingTablesMap[t] = true
	}

	for _, table := range tablesToClear {
		if !existingTablesMap[table] {
			g.Log().Debugf(ctx, "Table %s not found, skipping clear.", table)
			continue
		}
		_, err := db.Ctx(ctx).Delete(table)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to clear table %s: %v", table, err)
		}
		if db.GetConfig().DriverName == "sqlite" {
			_, errSeq := db.Ctx(ctx).Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", table))
			if errSeq != nil {
				g.Log().Debugf(ctx, "Failed to reset sqlite_sequence for table %s: %v", table, errSeq)
			}
		}
	}
	e, err := CasbinEnforcer()
	if err == nil && e != nil {
		e.ClearPolicy()
		if loadErr := e.LoadPolicy(); loadErr != nil { // Reload to ensure empty state from DB
			g.Log().Warningf(ctx, "Error reloading Casbin policy after table clear: %v", loadErr)
		}
	}
}

func createTestRole(ctx context.Context, code, name string) (uint64, error) {
	res, err := dao.SystemRole.Ctx(ctx).Data(g.Map{
		dao.SystemRole.Columns().Code:   code,
		dao.SystemRole.Columns().Name:   name,
		dao.SystemRole.Columns().Status: 0,
	}).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

func TestRoleService_CreateRole(t *testing.T) {
	ctx := testRoleCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestRoleAndRelatedTables(ctx)
		uniqueSuffix := grand.S(4)
		createInput := &input.RoleCreateInp{
			Name: "Test Role " + uniqueSuffix, Code: "test_role_" + uniqueSuffix, Status: 0, SortOrder: 10,
		}
		roleId, err := RoleService.CreateRole(ctx, createInput)
		t.AssertNil(err, "CreateRole should not error")
		t.AssertGT(roleId, 0, "CreateRole should return positive ID")
		role, _ := RoleService.GetRoleById(ctx, uint64(roleId))
		t.AssertNE(role, nil, "Role should exist after creation")
		t.AssertEQ(role.Name, createInput.Name)
	})
}

func TestRoleService_CreateRole_DuplicateCodeOrName(t *testing.T) {
	ctx := testRoleCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestRoleAndRelatedTables(ctx)
		uniqueSuffix := grand.S(4)
		roleCode := "dup_code_" + uniqueSuffix
		roleName := "Dup Name " + uniqueSuffix

		in1 := &input.RoleCreateInp{Name: roleName, Code: roleCode, Status: 0}
		_, err1 := RoleService.CreateRole(ctx, in1)
		t.AssertNil(err1, "First role creation should succeed")

		in2 := &input.RoleCreateInp{Name: "Another Name " + uniqueSuffix, Code: roleCode, Status: 0} // Same code
		_, err2 := RoleService.CreateRole(ctx, in2)
		t.AssertNE(err2, nil, "Should error on duplicate role code")
		t.Assert(gerror.Contains(err2, "already exists"), "Error message for duplicate code")

		in3 := &input.RoleCreateInp{Name: roleName, Code: "another_code_" + uniqueSuffix, Status: 0} // Same name
		_, err3 := RoleService.CreateRole(ctx, in3)
		t.AssertNE(err3, nil, "Should error on duplicate role name")
		t.Assert(gerror.Contains(err3, "already exists"), "Error message for duplicate name")
	})
}

func TestRoleService_UpdateRole(t *testing.T) {
	ctx := testRoleCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestRoleAndRelatedTables(ctx)
		uniqueSuffix := grand.S(4)
		roleId, _ := createTestRole(ctx, "orig_code_"+uniqueSuffix, "Original Name "+uniqueSuffix)

		updateInput := &input.RoleUpdateInp{
			Id: roleId, Name: "Updated Name " + uniqueSuffix, Code: "updated_code_" + uniqueSuffix, Status: 0, SortOrder: 5,
		}
		err := RoleService.UpdateRole(ctx, updateInput)
		t.AssertNil(err)
		updatedRole, _ := RoleService.GetRoleById(ctx, roleId)
		t.AssertEQ(updatedRole.Name, updateInput.Name)
		t.AssertEQ(updatedRole.Code, updateInput.Code)
	})
}

func TestRoleService_DeleteRole(t *testing.T) {
	ctx := testRoleCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestRoleAndRelatedTables(ctx)
		roleCode := "role_to_be_deleted"
		roleId, _ := createTestRole(ctx, roleCode, "Role To Delete")

		e, _ := CasbinEnforcer()
		_, _ = e.AddPolicy(roleCode, "/test/path", "GET") // p, roleCode, path, method

		// Create a dummy user and assign this role
		userInput := &input.UserCreateInp{Username: "user_with_role_to_delete", Password: "pw", Nickname: "Test", Status: 0}
		userId, _ := UserService.CreateUser(ctx, userInput) // UserService is in the same package
		userIdStr := gconv.String(userId)
		_, _ = e.AddGroupingPolicy(userIdStr, roleCode) // g, userId, roleCode
		// e.SavePolicy() // If adapter needs it

		err := RoleService.DeleteRole(ctx, []uint64{roleId})
		t.AssertNil(err)

		role, _ := RoleService.GetRoleById(ctx, roleId)
		t.AssertNil(role, "Role should be deleted from DB")

		// e.LoadPolicy() // Reload policies from DB
		t.Assert(!e.HasPolicy(roleCode, "/test/path", "GET"), "Casbin 'p' policy should be removed")
		t.Assert(!e.HasGroupingPolicy(userIdStr, roleCode), "Casbin 'g' policy should be removed")
	})
}

func TestRoleService_GetRoleList(t *testing.T) {
	ctx := testRoleCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestRoleAndRelatedTables(ctx)
		_, _ = createTestRole(ctx, "list_role_1", "List Role 1")
		_, _ = createTestRole(ctx, "list_role_2", "List Role 2")

		list, total, err := RoleService.GetRoleList(ctx, &input.RoleListInp{Page: 1, PageSize: 10})
		t.AssertNil(err)
		t.AssertEQ(total, 2)
		t.AssertEQ(len(list), 2)
	})
}

func TestRoleService_UpdateAndGetRolePermissions(t *testing.T) {
	ctx := testRoleCtxGlobal
	gtest.C(t, func(t *gtest.T) {
		clearTestRoleAndRelatedTables(ctx)
		roleCode := "perm_test_role"
		_, _ = createTestRole(ctx, roleCode, "Permissions Test Role")
		e, _ := CasbinEnforcer()

		permsToSet := []input.PermissionRule{
			{Path: "/api/data1", Method: "GET"}, {Path: "/api/data1", Method: "POST"},
		}
		err := RoleService.UpdateRolePermissions(ctx, &input.RoleSetPermissionsInp{RoleCode: roleCode, Permissions: permsToSet})
		t.AssertNil(err)

		// e.LoadPolicy() // Reload from DB
		retrievedPerms, _ := RoleService.GetRolePermissions(ctx, roleCode)
		t.AssertEQ(len(retrievedPerms), 2)
		t.Assert(e.HasPolicy(roleCode, "/api/data1", "GET"))
		t.Assert(e.HasPolicy(roleCode, "/api/data1", "POST"))

		// Update: remove one, add one
		permsToSet2 := []input.PermissionRule{
			{Path: "/api/data1", Method: "GET"}, {Path: "/api/data2", Method: "PUT"},
		}
		err = RoleService.UpdateRolePermissions(ctx, &input.RoleSetPermissionsInp{RoleCode: roleCode, Permissions: permsToSet2})
		t.AssertNil(err)
		// e.LoadPolicy()
		retrievedPerms2, _ := RoleService.GetRolePermissions(ctx, roleCode)
		t.AssertEQ(len(retrievedPerms2), 2)
		t.Assert(e.HasPolicy(roleCode, "/api/data1", "GET"))
		t.Assert(!e.HasPolicy(roleCode, "/api/data1", "POST"))
		t.Assert(e.HasPolicy(roleCode, "/api/data2", "PUT"))
	})
}
