package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"os"
	"testing"
	// "time" // Not directly used in this version of tests

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

var testCtxUserService = gctx.New() // Specific context for this test package

func TestMain(m *testing.M) {
	ctx := testCtxUserService
	projectRoot := gfile.normalize(gfile.Pwd() + "/../../../")

	configDirPath := gfile.Join(projectRoot, "manifest/config")
	configFileActualPath := gfile.Join(configDirPath, "config.yaml")

	if gfile.Exists(configFileActualPath) {
		if adapter, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
			adapter.SetPath(configDirPath)
			g.Log().Debugf(ctx, "TestMain (user_service_test): Set config path to: %s", configDirPath)
		} else {
			g.Log().Warningf(ctx, "TestMain (user_service_test): Config adapter is not *gcfg.AdapterFile, path not set.")
		}
	} else {
		g.Log().Warningf(ctx, "TestMain (user_service_test): config.yaml not found at: %s", configFileActualPath)
	}

	modelPathFromConfig := g.Cfg().MustGet(ctx, "casbin.modelPath", "manifest/config/casbin_model.conf").String()
	resolvedModelPath := modelPathFromConfig
	if !gfile.IsAbs(modelPathFromConfig) {
		if cfgAdapterPath := g.Cfg().GetAdapter().GetPath(); cfgAdapterPath != "" && gfile.Exists(gfile.Join(cfgAdapterPath, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(cfgAdapterPath, modelPathFromConfig)
		} else if projectRoot != "" && gfile.Exists(gfile.Join(projectRoot, modelPathFromConfig)) {
			resolvedModelPath = gfile.Join(projectRoot, modelPathFromConfig)
		} else if gfile.Exists(modelPathFromConfig) { // Check PWD if path is relative and not found via config/project root
			resolvedModelPath = modelPathFromConfig
		}
	}
	if !gfile.Exists(resolvedModelPath) {
		g.Log().Fatalf(ctx, "TestMain: Casbin model file '%s' not found. Tests cannot proceed.", resolvedModelPath)
	}
	g.Cfg().Set("casbin.modelPath", resolvedModelPath)
	g.Log().Infof(ctx, "TestMain (user_service_test): Effective Casbin model path: %s", resolvedModelPath)

	if _, err := CasbinEnforcer(); err != nil {
		g.Log().Fatalf(ctx, "TestMain: Failed to initialize Casbin enforcer: %v", err)
	}

	// Clean tables once before all tests in this package run
	clearTestUserAndRoleTables(ctx)

	exitCode := m.Run()
	os.Exit(exitCode)
}

func clearTestUserAndRoleTables(ctx context.Context) {
	tablesToClear := []string{
		dao.SystemUserRole.Table(),
		dao.SystemUserPost.Table(), // Assuming it might exist or be added
		dao.SystemUserDept.Table(), // Assuming it might exist or be added
		dao.SystemUser.Table(),
		dao.SystemRole.Table(),
		"casbin_rule", // Casbin table name
	}
	db := g.DB()
	for _, table := range tablesToClear {
		// Check if table exists before attempting to delete to avoid errors on non-existent tables
		// This is more robust if some tables are optional or managed elsewhere
		allTables, _ := db.Ctx(ctx).Tables()
		tableFound := false
		for _, t := range allTables {
			if t == table {
				tableFound = true
				break
			}
		}
		if !tableFound {
			g.Log().Debugf(ctx, "Table %s not found, skipping clear.", table)
			continue
		}

		_, err := db.Ctx(ctx).Delete(table)
		if err != nil {
			g.Log().Errorf(ctx, "Failed to clear table %s: %v", table, err)
		}
		if db.GetConfig().DriverName == "sqlite" {
			_, err = db.Ctx(ctx).Exec(fmt.Sprintf("DELETE FROM sqlite_sequence WHERE name='%s';", table))
			if err != nil {
				g.Log().Debugf(ctx, "Failed to reset sqlite_sequence for table %s (may not exist or error): %v", table, err)
			}
		}
	}
	e, err := CasbinEnforcer() // Get existing enforcer
	if err == nil && e != nil {
		e.ClearPolicy() // Clear in-memory policies
		// For some adapters, ClearPolicy doesn't affect DB. Direct delete was done above.
		// Reload to ensure in-memory matches (empty) DB state.
		if loadErr := e.LoadPolicy(); loadErr != nil {
			g.Log().Warningf(ctx, "Error reloading Casbin policy after table clear: %v", loadErr)
		}
	}
}

func createTestRoleForUserService(ctx context.Context, code, name string) (uint64, error) {
	res, err := dao.SystemRole.Ctx(ctx).Data(g.Map{
		dao.SystemRole.Columns().Code:   code,
		dao.SystemRole.Columns().Name:   name,
		dao.SystemRole.Columns().Status: 0, // Active
	}).Insert()
	if err != nil {
		return 0, err
	}
	id, _ := res.LastInsertId()
	return uint64(id), nil
}

// --- User Creation Tests ---
func TestUserService_CreateUser(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		uniqueSuffix := grand.S(4)
		createInput := &input.UserCreateInp{
			Username: "testuser_" + uniqueSuffix, Password: "password123", Nickname: "Test User " + uniqueSuffix,
			Email: "testuser_" + uniqueSuffix + "@example.com", Mobile: "1390000" + grand.N(4), Status: 0,
		}
		userId, err := UserService.CreateUser(ctx, createInput)
		t.AssertNil(err)
		t.AssertGT(userId, 0)
		user, _ := UserService.GetUserById(ctx, uint64(userId))
		t.AssertNE(user, nil)
		t.AssertEQ(user.Username, createInput.Username)
	})
}

func TestUserService_CreateUser_DuplicateUsername(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		createInput := &input.UserCreateInp{Username: "duplicate_user", Password: "p", Nickname: "D", Status: 0}
		UserService.CreateUser(ctx, createInput)           // First creation
		_, err := UserService.CreateUser(ctx, createInput) // Second attempt
		t.AssertNE(err, nil, "Should error on duplicate username")
		t.Assert(gerror.Contains(err, "already exists"), "Error message should indicate username exists")
	})
}

// --- User Login Tests ---
func TestUserService_UserLogin(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		password := "StrongP@ss1"
		userInput := &input.UserCreateInp{Username: "login_user", Password: password, Nickname: "L", Status: 0}
		UserService.CreateUser(ctx, userInput)

		user, err := UserService.UserLogin(ctx, &input.UserLoginInp{Username: userInput.Username, Password: password})
		t.AssertNil(err, "Login should succeed with correct credentials")
		t.AssertNE(user, nil, "Logged in user should not be nil")
		t.AssertEQ(user.Username, userInput.Username)
		t.AssertNE(user.LoginTime, nil, "LoginTime should be updated")
	})
}

func TestUserService_UserLogin_WrongPassword(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		userInput := &input.UserCreateInp{Username: "login_wrong_pw", Password: "password123", Nickname: "WP", Status: 0}
		UserService.CreateUser(ctx, userInput)
		_, err := UserService.UserLogin(ctx, &input.UserLoginInp{Username: userInput.Username, Password: "wrongpassword"})
		t.AssertNE(err, nil, "Login should fail with wrong password")
		t.Assert(gerror.Contains(err, "invalid username or password"), "Error message for wrong password")
	})
}

// --- GetUserById Test ---
func TestUserService_GetUserById(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		userInput := &input.UserCreateInp{Username: "getbyid_user", Password: "pw", Nickname: "Get", Status: 0}
		userId, _ := UserService.CreateUser(ctx, userInput)
		user, err := UserService.GetUserById(ctx, uint64(userId))
		t.AssertNil(err)
		t.AssertNE(user, nil)
		t.AssertEQ(user.Id, uint64(userId))

		_, errNotFound := UserService.GetUserById(ctx, 999999)
		t.AssertNE(errNotFound, nil, "Should return error for non-existent user")
		t.Assert(gerror.Contains(errNotFound, "not found"), "Error message for non-existent user")
	})
}

// --- UpdateUser Test ---
func TestUserService_UpdateUser(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		uniqueSuffix := grand.S(4)
		userInput := &input.UserCreateInp{
			Username: "updateuser_" + uniqueSuffix, Password: "pw", Nickname: "OrigNick", Status: 0, Email: "update_" + uniqueSuffix + "@example.com",
		}
		userId, _ := UserService.CreateUser(ctx, userInput)
		updateInput := &input.UserUpdateInp{
			Id: uint64(userId), Nickname: "UpdatedNick", Email: "updated_" + uniqueSuffix + "@example.com", Status: 0,
		}
		err := UserService.UpdateUser(ctx, updateInput)
		t.AssertNil(err)
		updatedUser, _ := UserService.GetUserById(ctx, uint64(userId))
		t.AssertEQ(updatedUser.Nickname, updateInput.Nickname)
		t.AssertEQ(updatedUser.Email, updateInput.Email)
	})
}

// --- DeleteUser Test (including Casbin policy check) ---
func TestUserService_DeleteUser(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)

		userInput := &input.UserCreateInp{Username: "delete_user_casbin", Password: "pw", Nickname: "Del", Status: 0}
		userId, _ := UserService.CreateUser(ctx, userInput)
		userIdStr := gconv.String(userId)

		roleCode := "role_for_delete_test"
		roleId, _ := createTestRoleForUserService(ctx, roleCode, "Role For Delete Test")

		assignRolesInput := &input.UserAssignRoleInp{UserId: uint64(userId), RoleIds: []uint64{roleId}}
		UserService.UpdateUserRoles(ctx, assignRolesInput) // Assign role, this also updates Casbin

		e, _ := CasbinEnforcer()
		hasPolicy := e.HasGroupingPolicy(userIdStr, roleCode)
		t.Assert(hasPolicy, "Casbin should have g(user,role) before delete")

		err := UserService.DeleteUser(ctx, uint64(userId))
		t.AssertNil(err, "DeleteUser should not error")

		deletedUser := &entity.SystemUser{}
		g.DB().Ctx(ctx).Model(dao.SystemUser.Table()).Where(dao.SystemUser.Columns().Id, userId).Unscoped().Scan(deletedUser)
		t.AssertNE(deletedUser.DeletedAt, nil, "DeletedAt should be set for soft delete")

		count, _ := dao.SystemUserRole.Ctx(ctx).Where(dao.SystemUserRole.Columns().UserId, userId).Count()
		t.AssertEQ(count, 0, "User-role link should be removed from DB")

		e.LoadPolicy() // Reload policies from DB to ensure it reflects changes
		hasPolicyAfterDelete := e.HasGroupingPolicy(userIdStr, roleCode)
		t.Assert(!hasPolicyAfterDelete, "Casbin g(user,role) should be removed after user delete")
	})
}

// --- ChangeUserPassword Test ---
func TestUserService_ChangeUserPassword(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)
		oldPassword := "oldPassword123"
		newPassword := "newPassword456"
		userInput := &input.UserCreateInp{Username: "changepw_user", Password: oldPassword, Nickname: "CPW", Status: 0}
		userId, _ := UserService.CreateUser(ctx, userInput)

		err := UserService.ChangeUserPassword(ctx, uint64(userId), oldPassword, newPassword)
		t.AssertNil(err)

		_, loginErr := UserService.UserLogin(ctx, &input.UserLoginInp{Username: userInput.Username, Password: newPassword})
		t.AssertNil(loginErr, "Login with new password should succeed")
	})
}

// --- UpdateUserRoles Test (including Casbin policy check) ---
func TestUserService_UpdateUserRoles(t *testing.T) {
	ctx := testCtxUserService
	gtest.C(t, func(t *gtest.T) {
		clearTestUserAndRoleTables(ctx)

		userInput := &input.UserCreateInp{Username: "urt_user", Password: "pw", Nickname: "URT", Status: 0}
		userId, _ := UserService.CreateUser(ctx, userInput)
		userIdStr := gconv.String(userId)

		role1Code, role1Id := "role_urt_1", uint64(0)
		role1Id, _ = createTestRoleForUserService(ctx, role1Code, "Role URT 1")

		role2Code, role2Id := "role_urt_2", uint64(0)
		role2Id, _ = createTestRoleForUserService(ctx, role2Code, "Role URT 2")

		role3Code, role3Id := "role_urt_3", uint64(0)
		role3Id, _ = createTestRoleForUserService(ctx, role3Code, "Role URT 3")

		e, _ := CasbinEnforcer()

		// Assign role1, role2
		assignInput1 := &input.UserAssignRoleInp{UserId: uint64(userId), RoleIds: []uint64{role1Id, role2Id}}
		err := UserService.UpdateUserRoles(ctx, assignInput1)
		t.AssertNil(err)
		e.LoadPolicy()
		t.Assert(e.HasGroupingPolicy(userIdStr, role1Code), "User should have role1")
		t.Assert(e.HasGroupingPolicy(userIdStr, role2Code), "User should have role2")

		// Update to role2, role3
		assignInput2 := &input.UserAssignRoleInp{UserId: uint64(userId), RoleIds: []uint64{role2Id, role3Id}}
		err = UserService.UpdateUserRoles(ctx, assignInput2)
		t.AssertNil(err)
		e.LoadPolicy()
		t.Assert(!e.HasGroupingPolicy(userIdStr, role1Code), "User should not have role1 anymore")
		t.Assert(e.HasGroupingPolicy(userIdStr, role2Code), "User should still have role2")
		t.Assert(e.HasGroupingPolicy(userIdStr, role3Code), "User should now have role3")

		// Update to remove all roles
		assignInput3 := &input.UserAssignRoleInp{UserId: uint64(userId), RoleIds: []uint64{}}
		err = UserService.UpdateUserRoles(ctx, assignInput3)
		t.AssertNil(err)
		e.LoadPolicy()
		t.Assert(!e.HasGroupingPolicy(userIdStr, role2Code), "User should not have role2 anymore")
		t.Assert(!e.HasGroupingPolicy(userIdStr, role3Code), "User should not have role3 anymore")
	})
}
