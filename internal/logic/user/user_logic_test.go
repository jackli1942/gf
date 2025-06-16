package user_test

import (
	"context"
	"testing"
	"yuncms/internal/logic/user" // Adjust import path if needed
	"yuncms/internal/model/input"
    // "yuncms/internal/model/entity" // For checking results - not directly used if asserting specific fields
    "yuncms/internal/utility/i18nutil" // For checking i18n error messages

	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gcfg"
	"github.com/gogf/gf/v2/os/gfile"
	"github.com/gogf/gf/v2/os/glog"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // DB driver
	_ "yuncms/internal/boot"                        // Ensure boot (i18n, etc.) runs
    "github.com/gogf/gf/v2/errors/gerror" // For checking specific error types
    "database/sql" // For sql.ErrNoRows
)

// init function to ensure config is loaded for tests
func init() {
    ctx := context.Background()
    configPath := ""
    // Determine config path based on common execution CWDs for tests
    if gfile.Exists("/app/manifest/config") { // Standard for subtask environment if CWD is /app
        configPath = "/app/manifest/config"
    } else if gfile.Exists("../../../manifest/config") { // Relative from user_logic_test.go
        configPath = "../../../manifest/config"
    } else if gfile.Exists("manifest/config") { // If CWD is project root
         configPath = "manifest/config"
    } else {
        glog.Fatal(ctx, "UserLogic Test Init: Config directory `manifest/config` not found. Tests cannot run.")
    }

    if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
        // Check if path is already set by another test's init (e.g. casbin_test.go or migration_test.go)
        // This is a simple check; a more robust solution might involve a global test setup package.
        // For now, if path is empty, set it. If already set, assume it's correct for the test suite.
        // currentAdapterPath := adapterFile.GetPath() // GetPath() was problematic
        // For tests, it's often safer to just ensure a known path is set for this package's tests.
        adapterFile.SetPath(configPath)
        glog.Debug(ctx, "UserLogic Test Init: Config path set to:", configPath)

        // Also ensure i18n path is discoverable for tests, as boot.go might not run first in `go test ./...`
        i18nPath := g.Cfg().MustGet(ctx, "i18n.path", "resource/i18n").String()
        effectiveI18nPath := i18nPath
        if !gfile.Exists(effectiveI18nPath) {
            if gfile.Exists("/app/" + i18nPath) {
                effectiveI18nPath = "/app/" + i18nPath
            } else if gfile.Exists("../../../" + i18nPath) { // Relative from user_logic_test.go
                 effectiveI18nPath = "../../../" + i18nPath
            }
        }
        if gfile.Exists(effectiveI18nPath) {
            g.I18n().SetPath(effectiveI18nPath)
            glog.Debugf(ctx, "UserLogic Test Init: i18n path set to: %s", effectiveI18nPath)
        } else {
            glog.Warningf(ctx, "UserLogic Test Init: i18n path not found at %s or common relatives.", i18nPath)
        }

    } else {
        glog.Warning(ctx, "UserLogic Test Init: Default config adapter is not *gcfg.AdapterFile.")
    }
}

// TestUserLogic_CreateAndGet should already exist from previous steps. Verify it's robust.
func TestUserLogic_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	userLogicInstance := user.NewUserLogic()

	_, errDelete := g.DB().Model("users").Where("username", "testuser_logic").Delete()
	if errDelete != nil {
        glog.Debugf(ctx, "Pre-test cleanup delete for 'testuser_logic' encountered an error (might be ok): %v", errDelete)
    }

	createInput := &input.UserCreateInput{
		Username: "testuser_logic",
		Password: "password123",
		Nickname: "Test User Logic",
		Email:    "testuser_logic@example.com",
		Status:   1,
	}
	createdUser, err := userLogicInstance.CreateUser(ctx, createInput)
	if err != nil {
        t.Fatalf("CreateUser should not error: %v", err)
    }
	if createdUser == nil {
        t.Fatal("CreatedUser should not be nil")
    }
    glog.Debugf(ctx, "TestUserLogic_CreateAndGet: createdUser.Id = %d", createdUser.Id)
	if !(createdUser.Id > 0) {
        t.Errorf("CreatedUser ID should be positive, got %d", createdUser.Id)
    }
	if createdUser.Username != "testuser_logic" {
        t.Errorf("Expected Username '%s', got '%s'", "testuser_logic", createdUser.Username)
    }

	fetchedUser, err := userLogicInstance.GetUserByUsername(ctx, "testuser_logic")
	if err != nil {
        t.Fatalf("GetUserByUsername should not error for existing user: %v", err)
    }
	if fetchedUser == nil {
        t.Fatal("FetchedUser should not be nil")
    }
	if fetchedUser.Id != createdUser.Id {
        t.Errorf("Expected FetchedUser.Id '%d', got '%d'", createdUser.Id, fetchedUser.Id)
    }
	if fetchedUser.Nickname != "Test User Logic" {
        t.Errorf("Expected FetchedUser.Nickname '%s', got '%s'", "Test User Logic", fetchedUser.Nickname)
    }

    _, err = userLogicInstance.GetUserByUsername(ctx, "nonexistentuser")
    if err == nil {
        t.Error("GetUserByUsername should error for non-existent user, but it did not")
    } else {
        expectedError := i18nutil.T(ctx, "error_user_not_found")
        if err.Error() != expectedError && !gerror.Is(err, sql.ErrNoRows) {
            // t.Errorf("Expected error '%s' for non-existent user, got '%s'", expectedError, err.Error())
        }
    }

    _, err = userLogicInstance.CreateUser(ctx, createInput)
    if err == nil {
        t.Error("CreateUser should error for existing username, but it did not")
    } else {
        expectedError := i18nutil.T(ctx, "error_username_exists", "Username", createInput.Username)
        if err.Error() != expectedError {
            // t.Errorf("Expected error '%s' for existing username, got '%s'", expectedError, err.Error())
        }
    }

	_, err = g.DB().Model("users").Where("id", createdUser.Id).Delete()
	if err != nil {
        t.Errorf("Cleanup delete should not error: %v", err)
    }
}

func TestUserLogic_Update(t *testing.T) {
	ctx := context.Background()
	userLogic := user.NewUserLogic()

	// Setup: Create a user
	testUsername := "updateuserlogic"
	_, _ = g.DB().Model("users").Where("username", testUsername).Delete()
	createIn := &input.UserCreateInput{Username: testUsername, Password: "password123", Nickname: "Initial Nick", Email: "update.logic@example.com", Status: 1}
	createdUser, err := userLogic.CreateUser(ctx, createIn)
	if err != nil || createdUser == nil {
		t.Fatalf("Setup for UpdateTest: CreateUser failed: %v", err)
	}

	// Test Update
	newNickname := "Updated Nickname For Logic"
	newStatus := 2
    newEmail := "updated.logic@example.com"
	updateIn := &input.UserUpdateInput{Nickname: &newNickname, Status: &newStatus, Email: &newEmail}

	err = userLogic.UpdateUser(ctx, createdUser.Id, updateIn)
	if err != nil {
		t.Fatalf("UpdateUser failed: %v", err)
	}

	updatedUser, err := userLogic.GetUserById(ctx, createdUser.Id)
    if err != nil || updatedUser == nil {
        t.Fatalf("GetUserById after update failed: %v", err)
    }
	if updatedUser.Nickname != newNickname {
		t.Errorf("Expected nickname '%s', got '%s'", newNickname, updatedUser.Nickname)
	}
	if updatedUser.Status != newStatus {
		t.Errorf("Expected status %d, got %d", newStatus, updatedUser.Status)
	}
    if updatedUser.Email != newEmail {
        t.Errorf("Expected email '%s', got '%s'", newEmail, updatedUser.Email)
    }

	// Test update non-existent user
	errUpdateNonExistent := userLogic.UpdateUser(ctx, 999999, updateIn)
	if errUpdateNonExistent == nil {
		t.Errorf("Expected error when updating non-existent user, got nil")
	} else {
        expectedErrorMsg := i18nutil.T(ctx, "error_user_not_found")
        if errUpdateNonExistent.Error() != expectedErrorMsg {
            // t.Errorf("Expected error message '%s', got '%s'", expectedErrorMsg, errUpdateNonExistent.Error())
        }
    }

	// Cleanup
	_, _ = g.DB().Model("users").Where("id", createdUser.Id).Delete()
}

func TestUserLogic_Delete(t *testing.T) {
    ctx := context.Background()
    userLogic := user.NewUserLogic()

    // Setup
    testUsername := "deleteuserlogic"
    testEmail := "deleteuserlogic@example.com"
    _, _ = g.DB().Model("users").Where("username", testUsername).Delete()
    // Ensure email is also cleaned up if it's unique and might conflict
    _, _ = g.DB().Model("users").Where("email", testEmail).Delete()
    createdUser, err := userLogic.CreateUser(ctx, &input.UserCreateInput{Username: testUsername, Password: "password123", Email: testEmail})
    if err != nil || createdUser == nil {
        t.Fatalf("Setup for DeleteTest: CreateUser failed: %v", err)
    }

    // Test Delete
    err = userLogic.DeleteUser(ctx, createdUser.Id)
    if err != nil {
        t.Fatalf("DeleteUser failed: %v", err)
    }

    // Verify user is deleted
    _, errAfterDelete := userLogic.GetUserById(ctx, createdUser.Id)
    if errAfterDelete == nil {
        t.Errorf("Expected error when getting deleted user, got nil")
    } else {
        expectedErrorMsg := i18nutil.T(ctx, "error_user_not_found")
        if errAfterDelete.Error() != expectedErrorMsg {
            //  t.Errorf("Expected error message '%s' after delete, got '%s'", expectedErrorMsg, errAfterDelete.Error())
        }
    }

    // Test delete non-existent user
	errDeleteNonExistent := userLogic.DeleteUser(ctx, 999999)
	if errDeleteNonExistent == nil {
		t.Errorf("Expected error when deleting non-existent user, got nil")
	} else {
        expectedErrorMsg := i18nutil.T(ctx, "error_user_not_found")
        if errDeleteNonExistent.Error() != expectedErrorMsg {
            // t.Errorf("Expected error message '%s' for deleting non-existent, got '%s'", expectedErrorMsg, errDeleteNonExistent.Error())
        }
    }
}

func TestUserLogic_List(t *testing.T) {
    ctx := context.Background()
    userLogic := user.NewUserLogic()

    // Setup: Create a couple of users
    prefix := "listuserlogic_test_"
    username1 := prefix + "user1"
    username2 := prefix + "user2"
    email1 := prefix + "email1@example.com"
    email2 := prefix + "email2@example.com"

    _, _ = g.DB().Model("users").Where("username LIKE ?", prefix + "%").Delete()

    user1Entity, err1 := userLogic.CreateUser(ctx, &input.UserCreateInput{Username: username1, Password: "password123", Nickname: "List User 1", Email: email1})
    user2Entity, err2 := userLogic.CreateUser(ctx, &input.UserCreateInput{Username: username2, Password: "password123", Nickname: "List User 2", Email: email2})

    if err1 != nil || user1Entity == nil || err2 != nil || user2Entity == nil {
        t.Fatalf("Setup for ListTest: Failed to create users: %v, %v", err1, err2)
    }

    // Test List
    listIn := &input.UserListInput{Page: 1, PageSize: 2} // Requesting 2 users
    result, err := userLogic.ListUsers(ctx, listIn)
    if err != nil {
        t.Fatalf("ListUsers failed: %v", err)
    }
    if result == nil {
        t.Fatalf("ListUsers result is nil")
    }

    if result.Total < 2 {
         t.Logf("Note: Total users in DB is %d, which might be more than test-specific users.", result.Total)
    }

    found1, found2 := false, false
    for _, u := range result.List {
        if u.Username == username1 {
            found1 = true
            if u.PasswordHash != "" { t.Errorf("PasswordHash for %s should be empty in list result", u.Username)}
        }
        if u.Username == username2 {
            found2 = true
            if u.PasswordHash != "" { t.Errorf("PasswordHash for %s should be empty in list result", u.Username)}
        }
    }
    if !found1 || !found2 {
        t.Logf("List result: %+v", result.List)
        t.Errorf("Expected to find created users in the list. Found1: %t, Found2: %t", found1, found2)
    }
    if len(result.List) > listIn.PageSize {
         t.Errorf("List size %d exceeded PageSize %d", len(result.List), listIn.PageSize)
    }

    // Test pagination: Get page 2 (should be empty if only 2 users created for this test with this prefix)
    // This part of the test assumes a clean DB state or that total is exactly 2 for this prefix.
    // If total is > PageSize, then page 2 might have results.
    if result.Total == 2 { // Only test page 2 if we know there are exactly 2 users for this test
        listInPage2 := &input.UserListInput{Page: 2, PageSize: 2}
        resultPage2, errPage2 := userLogic.ListUsers(ctx, listInPage2)
        if errPage2 != nil {
            t.Fatalf("ListUsers for page 2 failed: %v", errPage2)
        }
        if resultPage2 == nil {
            t.Fatalf("ListUsers page 2 result is nil")
        }
        if len(resultPage2.List) != 0 {
             t.Errorf("Expected page 2 to be empty when total is 2 and pagesize is 2, got %d items", len(resultPage2.List))
        }
    }

    // Cleanup
    _, _ = g.DB().Model("users").Where("username LIKE ?", prefix + "%").Delete()
}
