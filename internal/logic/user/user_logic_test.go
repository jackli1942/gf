package user_test

import (
	"context"
	"testing"
	"yuncms/internal/logic/user"
	"yuncms/internal/model/input"
	// "github.com/gogf/gf/v2/test/gtest" // No longer using gtest.C or gtest.Assert*
    "github.com/gogf/gf/v2/frame/g"
    "github.com/gogf/gf/v2/os/gcfg"
    "github.com/gogf/gf/v2/os/gfile"
    "github.com/gogf/gf/v2/os/glog"
    _ "github.com/gogf/gf/contrib/drivers/mysql/v2" // DB driver
)

func init() {
    ctx := context.Background()
    configPath := ""
    if gfile.Exists("/app/manifest/config") {
        configPath = "/app/manifest/config"
    } else if gfile.Exists("../../../manifest/config") {
        configPath = "../../../manifest/config"
    } else if gfile.Exists("manifest/config") {
         configPath = "manifest/config"
    } else {
        glog.Warning(ctx, "UserLogic Test Init: Config directory `manifest/config` not found via common paths.")
    }

    if configPath != "" {
        if adapterFile, ok := g.Cfg().GetAdapter().(*gcfg.AdapterFile); ok {
            adapterFile.SetPath(configPath)
            glog.Debug(ctx, "UserLogic Test Init: Config path set to:", configPath)
        } else {
            glog.Warning(ctx, "UserLogic Test Init: Default config adapter is not *gcfg.AdapterFile.")
        }
    }
}

func TestUserLogic_CreateAndGet(t *testing.T) {
	ctx := context.Background()
	userLogicInstance := user.NewUserLogic()

	// Clean up user if exists from previous failed run
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
    }

    _, err = userLogicInstance.CreateUser(ctx, createInput)
    if err == nil {
        t.Error("CreateUser should error for existing username, but it did not")
    }

	_, err = g.DB().Model("users").Where("id", createdUser.Id).Delete()
	if err != nil {
        t.Errorf("Cleanup delete should not error: %v", err)
    }
}
