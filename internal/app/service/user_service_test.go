package service_test

import (
	"context"
	"strings"
	"testing"
	"yuncms/internal/app/dao"
	"yuncms/internal/app/model"
	"yuncms/internal/app/service"

	"github.com/gogf/gf/v2/database/gdb"
	_ "github.com/gogf/gf/contrib/drivers/mysql/v2" // MySQL driver
	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/bcrypt"
)

const (
	testDBGroupUserService   = "user_service_test_db"
	testDBNameUserService    = "yuncms_dev_user_delete_test" // DB for User Delete tests
	testTableNameUserService = "users"
)

// setupTestDB configures a temporary database for testing.
func setupTestDB(t *testing.T) {
	gdb.AddConfigNode(gdb.DefaultGroupName, gdb.ConfigNode{
		Host:    "127.0.0.1",
		Port:    "3306",
		User:    "root",
		Pass:    "",
		Name:    testDBNameUserService,
		Type:    "mysql",
		Charset: "utf8mb4",
		Debug:   false,
	})

	db, err := gdb.Instance(gdb.DefaultGroupName)
	if err != nil {
		t.Fatalf("Failed to get DB instance for default group after override: %v", err)
	}

	err = db.PingMaster()
	if err != nil {
		t.Fatalf("Failed to ping DB for group %s: %v", gdb.DefaultGroupName, err) // Use gdb.DefaultGroupName here
	}
}

// cleanupUser removes a user by username for test cleanup.
func cleanupUser(ctx context.Context, t *testing.T, username string) {
	_, err := dao.User.M(ctx).Where(dao.User.Columns.Username, username).Delete()
	if err != nil {
		t.Logf("Cleanup failed for user %s: %v", username, err)
	}
}

func TestUserService_CreateUser(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	userService := service.NewUserService()

	t.Run("Successful Creation", func(t *testing.T) {
		statusVal := 1
		userInput := service.UserCreateInput{
			Username: "testuser_success", Password: "password123", Nickname: "Test User Success",
			Email: "success@example.com", Status: &statusVal,
		}
		defer cleanupUser(ctx, t, userInput.Username)
		createdUser, err := userService.CreateUser(ctx, userInput)
		assert.NoError(t, err)
		assert.NotNil(t, createdUser)
		assert.Positive(t, createdUser.Id)
		assert.Equal(t, userInput.Username, createdUser.Username)
		err = bcrypt.CompareHashAndPassword([]byte(createdUser.Password), []byte(userInput.Password))
		assert.NoError(t, err, "Hashed password should match original password")
	})

	t.Run("Duplicate Username", func(t *testing.T) {
		initialInput := service.UserCreateInput{Username: "testuser_duplicate", Password: "password123", Nickname: "Test User Duplicate", Email: "duplicate1@example.com"}
		defer cleanupUser(ctx, t, initialInput.Username)
		_, err := userService.CreateUser(ctx, initialInput)
		assert.NoError(t, err)
		duplicateInput := service.UserCreateInput{Username: "testuser_duplicate", Password: "password456", Nickname: "Another User", Email: "duplicate2@example.com"}
		createdUser, err := userService.CreateUser(ctx, duplicateInput)
		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.True(t, strings.Contains(err.Error(), "already exists"))
	})

	t.Run("Duplicate Email", func(t *testing.T) {
		sharedEmail := "shared_email_create@example.com" // Use different email from update tests
		userInput1 := service.UserCreateInput{Username: "user_email1_create", Password: "password123", Nickname: "User Email 1", Email: sharedEmail}
		defer cleanupUser(ctx, t, userInput1.Username)
		_, err := userService.CreateUser(ctx, userInput1)
		assert.NoError(t, err)
		userInput2 := service.UserCreateInput{Username: "user_email2_create", Password: "password456", Nickname: "User Email 2", Email: sharedEmail}
		defer cleanupUser(ctx, t, userInput2.Username)
		createdUser, err := userService.CreateUser(ctx, userInput2)
		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.True(t, strings.Contains(err.Error(), "Email '"+sharedEmail+"' already exists"))
	})

	t.Run("Input Validation Failure - Missing Username", func(t *testing.T) {
		invalidInput := service.UserCreateInput{Password: "password123", Nickname: "Test User Invalid"}
		createdUser, err := userService.CreateUser(ctx, invalidInput)
		assert.Error(t, err)
		assert.Nil(t, createdUser)
		assert.True(t, strings.Contains(gerror.Cause(err).Error(), "Username is required"))
	})

	// Add other CreateUser validation sub-tests similarly...
}

// Seeded user data for tests.
var (
	seedUser1 = model.User{Id: 1, Uuid: "uuid-update-test-1", Username: "testuser_update", Nickname: "Initial Nickname", Email: "initial@example.com", Status: 1}
	seedUser2 = model.User{Id: 2, Uuid: "uuid-update-test-2", Username: "otheruser_update", Nickname: "Other User", Email: "other@example.com", Status: 1}
)

// Helper to get a string pointer
func strPtr(s string) *string { return &s }

// Helper to get an int pointer
func intPtr(i int) *int { return &i }

func TestUserService_GetUserByID(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	userService := service.NewUserService()
	t.Run("Fetch Existing User by ID", func(t *testing.T) {
		user, err := userService.GetUserByID(ctx, seedUser1.Id)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, seedUser1.Username, user.Username)
	})
	t.Run("Fetch Non-existent User by ID", func(t *testing.T) {
		user, err := userService.GetUserByID(ctx, 99999)
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_GetUserByUsername(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	userService := service.NewUserService()
	t.Run("Fetch Existing User by Username", func(t *testing.T) {
		user, err := userService.GetUserByUsername(ctx, seedUser2.Username)
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, seedUser2.Id, user.Id)
	})
	t.Run("Fetch Non-existent User by Username", func(t *testing.T) {
		user, err := userService.GetUserByUsername(ctx, "nonexistentuser")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_GetUserByUuid(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	userService := service.NewUserService()
	t.Run("Fetch Existing User by UUID", func(t *testing.T) {
		user, err := userService.GetUserByUuid(ctx, seedUser1.Uuid) // Use seedUser1 as seedUser3 is removed
		assert.NoError(t, err)
		assert.NotNil(t, user)
		assert.Equal(t, seedUser1.Id, user.Id)
	})
	t.Run("Fetch Non-existent User by UUID", func(t *testing.T) {
		user, err := userService.GetUserByUuid(ctx, "nonexistent-uuid-123")
		assert.NoError(t, err)
		assert.Nil(t, user)
	})
}

func TestUserService_ListUsers(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	userService := service.NewUserService()
	const totalSeededUsers = 2 // Only 2 users are seeded for this subtask's DB

	t.Run("List All Users (Page 1, Size covering all)", func(t *testing.T) {
		input := service.ListUsersInput{Page: 1, Size: 10}
		output, err := userService.ListUsers(ctx, input)
		assert.NoError(t, err)
		assert.NotNil(t, output)
		assert.Equal(t, totalSeededUsers, output.Total)
		assert.Len(t, output.Items, totalSeededUsers)
	})
	// Add other ListUsers sub-tests similarly...
}

func TestUserService_UpdateUser(t *testing.T) {
	setupTestDB(t)
	ctx := context.Background()
	userService := service.NewUserService()
	const existingUserID uint = 1
	const nonExistentUserID uint = 9999

	t.Run("Successful Nickname Update", func(t *testing.T) {
		newNickname := "Updated Nickname"
		input := service.UserUpdateInput{Nickname: strPtr(newNickname)}
		updatedUser, err := userService.UpdateUser(ctx, existingUserID, input)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newNickname, updatedUser.Nickname)
		originalEmailFromSeed := "initial@example.com" // User 1's seeded email
		assert.Equal(t, originalEmailFromSeed, updatedUser.Email)
	})

	t.Run("Successful Email Update", func(t *testing.T) {
		newEmail := "updated_email@example.com"
		input := service.UserUpdateInput{Email: strPtr(newEmail)}
		originalUser, errDb := dao.User.GetById(ctx, existingUserID)
		assert.NoError(t, errDb)
		assert.NotNil(t, originalUser, "Original user for email update test should exist")

		updatedUser, err := userService.UpdateUser(ctx, existingUserID, input)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		assert.Equal(t, newEmail, updatedUser.Email)

		if originalUser != nil {
			revertErr := dao.User.Update(ctx, existingUserID, gdb.Map{
				dao.User.Columns.Email:     originalUser.Email,
				dao.User.Columns.UpdatedAt: gtime.Now(),
			})
			assert.NoError(t, revertErr, "Failed to revert email after test")
		}
	})

	t.Run("Successful Password Update", func(t *testing.T) {
		newPassword := "newSecurePassword123"
		input := service.UserUpdateInput{Password: strPtr(newPassword)}
		updatedUser, err := userService.UpdateUser(ctx, existingUserID, input)
		assert.NoError(t, err)
		assert.NotNil(t, updatedUser)
		err = bcrypt.CompareHashAndPassword([]byte(updatedUser.Password), []byte(newPassword))
		assert.NoError(t, err, "New hashed password should match")
	})

	t.Run("Update Non-existent User", func(t *testing.T) {
		input := service.UserUpdateInput{Nickname: strPtr("No Such User")}
		_, err := userService.UpdateUser(ctx, nonExistentUserID, input)
		assert.Error(t, err)
		assert.True(t, gerror.Code(err) == gcode.CodeNotFound, "Error code should be CodeNotFound")
	})

	t.Run("Update Email to Existing (taken by another user)", func(t *testing.T) {
		input := service.UserUpdateInput{Email: strPtr(seedUser2.Email)} // seedUser2's email
		_, err := userService.UpdateUser(ctx, existingUserID, input)
		assert.Error(t, err)
		assert.True(t, gerror.Code(err) == gcode.CodeBusinessValidationFailed, "Error code should be CodeBusinessValidationFailed")
		assert.True(t, strings.Contains(err.Error(), "email_already_taken"))
	})

	// Add other UpdateUser validation/uniqueness sub-tests similarly...
}


func TestUserService_DeleteUser(t *testing.T) {
	setupTestDB(t) // Configures DB to yuncms_dev_user_delete_test
	ctx := context.Background()
	userService := service.NewUserService()

	// User ID 1 ('delete_me') and ID 2 ('survivor') are seeded by the bash script
	// seedUser1 corresponds to 'testuser_update' (ID 1 in general test setup) -> for delete, this is 'delete_me'
	// seedUser2 corresponds to 'otheruser_update' (ID 2 in general test setup) -> for delete, this is 'survivor'

	// For clarity in this test, let's use the specific IDs from the seed script for this subtask
	const userToDeleteID uint = 1    // username 'delete_me'
	const userToSurviveID uint = 2   // username 'survivor'
	const nonExistentUserIDForDelete uint = 99988
	stringUserToDeleteID := "1" // For Casbin

	// Get the Casbin enforcer instance
	// Note: For robust Casbin testing, ensuring it uses the test DB is crucial.
	// The current service.Casbin() initializes with config values. If these are not
	// overridden for the test scope, Casbin operations might target the main DB.
	// This test assumes service.Casbin() is correctly configured or its interaction is being checked at a high level.
	casbinEnforcer := service.Casbin()

	// Add some Casbin rules for the user to be deleted, to test cleanup
	// These rules will be added to whichever DB Casbin is currently configured for.
	casbinEnforcer.AddPolicy(stringUserToDeleteID, "/test/resource", "read")
	casbinEnforcer.AddGroupingPolicy(stringUserToDeleteID, "test_role")
	casbinEnforcer.SavePolicy() // Ensure policies are saved if using a persistent adapter

	t.Run("Successful Deletion", func(t *testing.T) {
		// Ensure user to delete exists before deletion
		userBeforeDelete, _ := userService.GetUserByID(ctx, userToDeleteID)
		assert.NotNil(t, userBeforeDelete, "User to delete should exist before deletion")

		err := userService.DeleteUser(ctx, userToDeleteID)
		assert.NoError(t, err)

		// Verify user is deleted from DB
		deletedUser, errDb := userService.GetUserByID(ctx, userToDeleteID)
		assert.NoError(t, errDb)
		assert.Nil(t, deletedUser, "User should be deleted from DB")

		// Verify Casbin rules are removed
		rolesAfterDelete, _ := casbinEnforcer.GetRolesForUser(stringUserToDeleteID)
		assert.Empty(t, rolesAfterDelete, "Casbin roles for deleted user should be empty")

		permissionsAfterDelete := casbinEnforcer.GetPermissionsForUser(stringUserToDeleteID)
		assert.Empty(t, permissionsAfterDelete, "Casbin permissions for deleted user should be empty")

		// Verify survivor user still exists
		survivorUser, _ := userService.GetUserByID(ctx, userToSurviveID)
		assert.NotNil(t, survivorUser, "Survivor user should still exist")
		if survivorUser != nil { // Defensive check
			assert.Equal(t, "survivor", survivorUser.Username)
		}
	})

	t.Run("Delete Non-existent User", func(t *testing.T) {
		err := userService.DeleteUser(ctx, nonExistentUserIDForDelete)
		assert.Error(t, err, "Deleting a non-existent user should return an error")
		assert.True(t, gerror.Code(err) == gcode.CodeNotFound, "Error code should be CodeNotFound for non-existent user deletion")
	})

	// Cleanup: Remove any rules added for 'survivor' if tests were more complex,
	// or clear Casbin policies for a truly clean state if tests run sequentially affecting global Casbin.
	// For this test, 'survivor' Casbin rules were not added.
	// If other tests added rules for user "2" (survivor's ID), they might need cleanup.
	// The AddPolicy and AddGroupingPolicy for stringUserToDeleteID are cleaned by DeleteUser.
}
