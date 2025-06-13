package service

import (
	"context"
	"yuncms/internal/app/dao"
	"yuncms/internal/app/model"

	"github.com/gogf/gf/v2/errors/gcode" // Added import for gcode
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv" // Added for gconv.String()
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserCreateInput is the DTO for creating a new user.
type UserCreateInput struct {
	Username string `json:"username" v:"required|length:3,30#user.usernameRequired|user.usernameLength"`
	Password string `json:"password" v:"required|length:6,30#user.passwordRequired|user.passwordLength"`
	Nickname string `json:"nickname" v:"required|length:1,30#user.nicknameRequired|user.nicknameLength"`
	Email    string `json:"email"    v:"email#user.emailFormat"`
	Phone    string `json:"phone"    v:"phone-loose#user.phoneFormat"`
	Avatar   string `json:"avatar"   v:"url#user.avatarUrl"`
	Status   *int   `json:"status"   v:"in:0,1#user.statusInvalid"`
	RoleIds  []uint `json:"roleIds,omitempty" description:"List of role IDs to assign to the user"`
}

// UserService handles user-related business logic.
type UserService struct{}

// NewUserService creates and returns a new UserService.
func NewUserService() *UserService {
	return &UserService{}
}

// CreateUser handles the logic for creating a new user.
func (s *UserService) CreateUser(ctx context.Context, in UserCreateInput) (*model.User, error) {
	// 1. Validate input
	validationErr := g.Validator().Data(in).Run(ctx)
	if validationErr != nil {
		return nil, validationErr // validationErr is already a gerror or compatible
	}

	// 2. Check if username already exists
	count, dbErr := dao.User.M(ctx).Where(dao.User.Columns.Username, in.Username).Count()
	if dbErr != nil {
		return nil, gerror.Wrap(dbErr, "Failed to check username existence")
	}
	if count > 0 {
		return nil, gerror.Newf("Username '%s' already exists", in.Username)
	}

	// 3. Check if email already exists (if provided)
	if in.Email != "" {
		count, dbErr = dao.User.M(ctx).Where(dao.User.Columns.Email, in.Email).Count()
		if dbErr != nil {
			return nil, gerror.Wrap(dbErr, "Failed to check email existence")
		}
		if count > 0 {
			return nil, gerror.Newf("Email '%s' already exists", in.Email)
		}
	}


	// 4. Hash password
	hashedPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if bcryptErr != nil {
		return nil, gerror.Wrap(bcryptErr, "Failed to hash password")
	}

	// 5. Populate user entity
	now := gtime.Now()
	userEntity := &model.User{
		Uuid:      uuid.NewString(),
		Username:  in.Username,
		Password:  string(hashedPassword),
		Nickname:  in.Nickname,
		Email:     in.Email,
		Phone:     in.Phone,
		Avatar:    in.Avatar,
		// Status handled below
		CreatedAt: now,
		UpdatedAt: now,
	}

	// Default status to 1 (active) if not provided or nil
	if in.Status == nil {
		userEntity.Status = 1
	} else {
		userEntity.Status = *in.Status
	}

	// 6. Call DAO to create user
	lastInsertId, createErr := dao.User.Create(ctx, userEntity)
	if createErr != nil {
		return nil, gerror.Wrap(createErr, "Failed to create user in database")
	}

	// 7. Set the ID on the entity and return
	userEntity.Id = uint(lastInsertId)
	// Password should not be returned, it's already `json:"-"` in the entity
	userEntity.Password = "" // Explicitly clear for return DTO consistency

	// 8. Assign roles via Casbin
	if len(in.RoleIds) > 0 {
		e := Casbin() // Get Casbin enforcer
		if e == nil {
			g.Log().Error(ctx, "CreateUser: Casbin enforcer is nil, cannot assign roles.")
			// Depending on policy, you might return an error here or just log
		} else {
			userIdStr := gconv.String(userEntity.Id)
			for _, roleId := range in.RoleIds {
				roleIdStr := gconv.String(roleId) // Assuming roles are identified by their stringified ID in Casbin
				// Check if role exists can be added here if necessary by querying a role table
				_, err := e.AddGroupingPolicy(userIdStr, roleIdStr)
				if err != nil {
					g.Log().Errorf(ctx, "CreateUser: Failed to assign role ID %s to user ID %s: %v", roleIdStr, userIdStr, err)
					// Decide on error handling: continue, or collect errors, or return immediately
				} else {
					g.Log().Debugf(ctx, "CreateUser: Successfully assigned role ID %s to user ID %s", roleIdStr, userIdStr)
				}
			}
		}
	}
	return userEntity, nil
}

// GetUserByID retrieves a user by their ID.
// Returns nil, nil if user not found, or user and error for other issues.
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*model.User, error) {
	user, err := dao.User.GetById(ctx, id)
	if err != nil {
		// Check if the error is due to "record not found"
		// gdb.ErrNotFound is not directly exposed, so check if user is nil
		if user == nil {
			return nil, nil // Standard practice: return nil, nil for not found
		}
		return nil, gerror.Wrapf(err, "Failed to get user by ID: %d", id)
	}
	return user, nil
}

// GetUserByUsername retrieves a user by their username.
// Returns nil, nil if user not found.
func (s *UserService) GetUserByUsername(ctx context.Context, username string) (*model.User, error) {
	user, err := dao.User.GetByUsername(ctx, username)
	if err != nil {
		if user == nil {
			return nil, nil
		}
		return nil, gerror.Wrapf(err, "Failed to get user by username: %s", username)
	}
	return user, nil
}

// GetUserByUuid retrieves a user by their UUID.
// Returns nil, nil if user not found.
func (s *UserService) GetUserByUuid(ctx context.Context, uuid string) (*model.User, error) {
	user, err := dao.User.GetByUuid(ctx, uuid)
	if err != nil {
		if user == nil {
			return nil, nil
		}
		return nil, gerror.Wrapf(err, "Failed to get user by UUID: %s", uuid)
	}
	return user, nil
}

// ListUsersInput defines the input for listing users with pagination.
type ListUsersInput struct {
	Page int `json:"page" v:"min:1#Page number must be positive" default:"1"`
	Size int `json:"size" v:"min:1|max:100#Page size must be between 1 and 100" default:"10"`
}

// ListUsersOutput defines the output for listing users.
type ListUsersOutput struct {
	Items []*model.User `json:"items"`
	Total int           `json:"total"`
	Page  int           `json:"page"`
	Size  int           `json:"size"`
}

// UserUpdateInput is the DTO for updating an existing user.
// Fields are pointers to distinguish between zero-values and fields not provided.
type UserUpdateInput struct {
	Nickname  *string `json:"nickname" v:"length:1,30#user.nicknameLength"`
	Password  *string `json:"password" v:"length:6,30#user.passwordLength"` // New password
	Email     *string `json:"email"    v:"email#user.emailFormat"`
	Phone     *string `json:"phone"    v:"phone-loose#user.phoneFormat"`
	Avatar    *string `json:"avatar"   v:"url#user.avatarUrl"`
	Status    *int    `json:"status"   v:"in:0,1#user.statusInvalid"`
	RoleIds   *[]uint `json:"roleIds,omitempty" description:"List of role IDs. If provided, existing roles are replaced. If nil, roles are not changed."`
}

// ListUsers retrieves a paginated list of users.
func (s *UserService) ListUsers(ctx context.Context, in ListUsersInput) (*ListUsersOutput, error) {
	// Validate input
	if err := g.Validator().Data(in).Run(ctx); err != nil {
		return nil, err
	}

	offset := (in.Page - 1) * in.Size

	users, err := dao.User.List(ctx, offset, in.Size)
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to list users from database")
	}
	if users == nil { // Ensure users is an empty slice, not nil, if no records found
		users = []*model.User{}
	}
	for _, u := range users { // Clear passwords for all users in list
		u.Password = ""
	}

	total, err := dao.User.Count(ctx)
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to count users from database")
	}

	return &ListUsersOutput{
		Items: users,
		Total: total,
		Page:  in.Page,
		Size:  in.Size,
	}, nil
}

// UpdateUser handles the logic for updating an existing user.
func (s *UserService) UpdateUser(ctx context.Context, id uint, in UserUpdateInput) (*model.User, error) {
	// 1. Fetch existing user
	currentUser, err := dao.User.GetById(ctx, id)
	if err != nil {
		return nil, gerror.Wrapf(err, "Error fetching user with ID %d", id)
	}
	if currentUser == nil {
		return nil, gerror.NewCodef(gcode.CodeNotFound, "User with ID %d not found", id)
	}

	// 2. Validate input
	// Note: Validator runs on all fields in `in`, even if some are nil.
	// The validation rules should be appropriate for optional fields (e.g. `v:"length:1,30"` on a *string will validate if not nil).
	if err := g.Validator().Data(in).Run(ctx); err != nil { // Validates UserUpdateInput including RoleIds if tags were there
		return nil, err
	}

	updateData := g.Map{}

	// 3. Process Nickname
	if in.Nickname != nil {
		if *in.Nickname != currentUser.Nickname {
			updateData[dao.User.Columns.Nickname] = *in.Nickname
		}
	}

	// 4. Process Email (check for uniqueness if changed)
	if in.Email != nil {
		if *in.Email != currentUser.Email {
			existingUser, err := dao.User.GetByEmailExcludeId(ctx, *in.Email, id)
			if err != nil {
				return nil, gerror.Wrapf(err, "Error checking email uniqueness for %s", *in.Email)
			}
			if existingUser != nil {
				// Use i18n key here, or a more structured error
				return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, "email_already_taken")
			}
			updateData[dao.User.Columns.Email] = *in.Email
		}
	}

	// 5. Process Phone (check for uniqueness if changed and if phone is meant to be unique)
	// Assuming phone uniqueness is required for this example.
	if in.Phone != nil {
		if *in.Phone != currentUser.Phone {
			existingUser, err := dao.User.GetByPhoneExcludeId(ctx, *in.Phone, id)
			if err != nil {
				return nil, gerror.Wrapf(err, "Error checking phone uniqueness for %s", *in.Phone)
			}
			if existingUser != nil {
				return nil, gerror.NewCode(gcode.CodeBusinessValidationFailed, "phone_already_taken")
			}
			updateData[dao.User.Columns.Phone] = *in.Phone
		}
	}

	// 6. Process Password
	if in.Password != nil && *in.Password != "" { // Ensure password is not empty string if provided
		hashedPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(*in.Password), bcrypt.DefaultCost)
		if bcryptErr != nil {
			return nil, gerror.Wrap(bcryptErr, "Failed to hash new password")
		}
		updateData[dao.User.Columns.Password] = string(hashedPassword)
	}

	// 7. Process Avatar
	if in.Avatar != nil {
		if *in.Avatar != currentUser.Avatar {
			updateData[dao.User.Columns.Avatar] = *in.Avatar
		}
	}

	// 8. Process Status
	if in.Status != nil {
		if *in.Status != currentUser.Status {
			updateData[dao.User.Columns.Status] = *in.Status
		}
	}

	// 9. If no fields to update, return current user or specific message
	if len(updateData) == 0 {
		return currentUser, nil // Or return an error/message indicating no changes were made
	}

	// Add UpdatedAt timestamp
	updateData[dao.User.Columns.UpdatedAt] = gtime.Now()

	// 10. Call DAO to update
	err = dao.User.Update(ctx, id, updateData)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to update user with ID %d", id)
	}

	// 11. Retrieve and return updated user
	updatedUser, err := dao.User.GetById(ctx, id)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to retrieve updated user with ID %d", id)
	}
	if updatedUser == nil { // Should not happen if update succeeded
		return nil, gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "user.notFoundId", id))
	}
	updatedUser.Password = "" // Clear password for return

	// 12. Update Casbin roles if RoleIds field is provided
	if in.RoleIds != nil {
		e := Casbin()
		if e == nil {
			g.Log().Error(ctx, "UpdateUser: Casbin enforcer is nil, cannot update roles.")
			// Depending on policy, might return error or just log
		} else {
			userIdStr := gconv.String(id)
			// Remove existing roles for the user
			_, err = e.RemoveFilteredGroupingPolicy(0, userIdStr)
			if err != nil {
				g.Log().Errorf(ctx, "UpdateUser: Failed to remove existing roles for user ID %s: %v", userIdStr, err)
				// Decide on error handling
			} else {
				g.Log().Debugf(ctx, "UpdateUser: Successfully removed existing roles for user ID %s", userIdStr)
			}

			// Add new roles
			for _, roleId := range *in.RoleIds {
				roleIdStr := gconv.String(roleId)
				// Check if role exists can be added here
				added, err := e.AddGroupingPolicy(userIdStr, roleIdStr)
				if err != nil {
					g.Log().Errorf(ctx, "UpdateUser: Failed to add role ID %s to user ID %s: %v", roleIdStr, userIdStr, err)
				} else if !added {
					g.Log().Warningf(ctx, "UpdateUser: Policy for role ID %s to user ID %s may already exist or was not added.", roleIdStr, userIdStr)
				} else {
					g.Log().Debugf(ctx, "UpdateUser: Successfully added role ID %s to user ID %s", roleIdStr, userIdStr)
				}
			}
		}
	}

	return updatedUser, nil
}

// DeleteUser handles the logic for deleting a user by ID.
// It also cleans up associated Casbin rules.
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	// Attempt to delete user from database
	err := dao.User.Delete(ctx, id)
	if err != nil {
		// If DAO returns gcode.NotFound, propagate it. Otherwise, wrap.
		if gerror.Code(err) == gcode.CodeNotFound {
			return err // Return the specific "not found" error from DAO
		}
		return gerror.Wrapf(err, "Failed to delete user with ID %d from database", id)
	}

	// If database deletion was successful, proceed to clean up Casbin rules
	// Convert ID to string as Casbin typically uses string subjects
	stringId := g.NewVar(id).String()

	// DeleteUser from Casbin removes all policies and grouping policies associated with the user.
	// The dobyte/gf-casbin Enforcer's DeleteUser method should handle this.
	// Note: service.Casbin() initializes and returns the global enforcer.
	// casbin_service.go is now in the same package directory, Casbin() should be directly accessible.
	if casbinEnforcer := Casbin(); casbinEnforcer != nil {
		// We expect DeleteUser to return (bool, error) or just error depending on the Casbin API version/adapter
		// Assuming dobytecasbin.Enforcer.DeleteUser returns (bool, error) similar to casbin.Enforcer
		deletedCasbinUser, casbinErr := casbinEnforcer.DeleteUser(stringId)
		if casbinErr != nil {
			g.Log().Warningf(ctx, "Error while deleting user '%s' from Casbin: %v. Policies might not have been fully removed.", stringId, casbinErr)
			// Do not return this error as the primary DB operation was successful.
			// This should be handled via logging or a more sophisticated error aggregation if needed.
		} else if !deletedCasbinUser {
			g.Log().Warningf(ctx, "User '%s' not found or no rules to delete in Casbin. This might be normal if user had no specific rules.", stringId)
		} else {
			g.Log().Infof(ctx, "Successfully deleted user '%s' and their associated rules from Casbin.", stringId)
		}
	} else {
		g.Log().Error(ctx, "Failed to get Casbin enforcer instance for user rule cleanup.")
		// This indicates a problem with Casbin initialization itself.
	}

	return nil // User deleted successfully from DB
}

// --- User Login ---

// UserLoginInput is the DTO for user login.
type UserLoginInput struct {
	Username string `json:"username" v:"required#auth.usernameRequired"`
	Password string `json:"password" v:"required#auth.passwordRequired"`
}

// UserLoginOutput is the DTO for user login response.
type UserLoginOutput struct {
	Token    string      `json:"token"`
	ExpireAt int64       `json:"expireAt"` // Expiration timestamp for the token
	User     *model.User `json:"user"`
}

// LoginUser handles the user login process.
func (s *UserService) LoginUser(ctx context.Context, in UserLoginInput) (out *UserLoginOutput, err error) {
	// 1. Validate input
	if errVal := g.Validator().Data(in).Run(ctx); errVal != nil {
		// Use the validation messages from struct tags (e.g., #auth.usernameRequired)
		return nil, gerror.WrapCode(gcode.CodeValidationFailed, errVal)
	}

	// 2. Fetch user by username
	user, err := dao.User.GetByUsername(ctx, in.Username) // Assumes this DAO method exists
	if err != nil {
		g.Log().Errorf(ctx, "LoginUser: Error fetching user '%s': %v", in.Username, err)
		return nil, gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.invalidCredentials"))
	}
	if user == nil {
		return nil, gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.invalidCredentials"))
	}

	// 3. Check user status
	if user.Status == 0 { // Assuming 0 means disabled
		return nil, gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.userDisabled"))
	}

	// 4. Compare hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(in.Password))
	if err != nil { // Handles mismatch (ErrMismatchedHashAndPassword) and other potential errors
		return nil, gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.invalidCredentials"))
	}

	// 5. Generate JWT token using TokenService
	// Ensure NewTokenService() returns ITokenService from internal/service (not internal/app/service)
	tokenGenService := NewTokenService()
	tokenString, expireAt, tokenErr := tokenGenService.GenerateUserToken(ctx, user.Id, user.Username)

	if tokenErr != nil {
		// token_service.GenerateUserToken already logs and wraps error with i18n key "auth.tokenGenerationFailed"
		return nil, tokenErr
	}

	// 6. Prepare output
	user.Password = "" // Clear password before returning

	out = &UserLoginOutput{
		Token:    tokenString,
		ExpireAt: expireAt,
		User:     user,
	}

	return out, nil
}

// InitPassword allows an administrator to reset/initialize a user's password.
func (s *UserService) InitPassword(ctx context.Context, userId uint, newPassword string) error {
	// 1. Validate userId (basic check, controller should also validate path param)
	if userId == 0 {
		return gerror.NewCodef(gcode.CodeInvalidArgument, g.I18n().T(ctx, "user.idInvalid"))
	}

	// 2. Check if user exists
	user, err := dao.User.GetById(ctx, userId)
	if err != nil {
		return gerror.Wrapf(err, "InitPassword: Failed to retrieve user with ID %d", userId)
	}
	if user == nil {
		return gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "user.notFoundId", userId))
	}

	// 3. Validate newPassword length (service level validation as well)
	// This duplicates controller DTO validation but ensures service integrity.
	// Using a generic validation error key here.
	if len(newPassword) < 6 || len(newPassword) > 30 { // Assuming same constraints as create/update
		return gerror.NewCodef(gcode.CodeValidationFailed, g.I18n().T(ctx, "validation.passwordLength"))
	}

	// 4. Hash the new password
	hashedPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if bcryptErr != nil {
		g.Log().Errorf(ctx, "InitPassword: Failed to hash new password for UserID %d: %v", userId, bcryptErr)
		return gerror.Wrap(bcryptErr, "Failed to hash new password")
	}

	// 5. Update the user's password in the database
	updateData := g.Map{
		dao.User.Columns.Password:  string(hashedPassword),
		dao.User.Columns.UpdatedAt: gtime.Now(),
	}

	_, err = dao.User.Ctx(ctx).Data(updateData).Where(dao.User.Columns.Id, userId).Update()
	if err != nil {
		g.Log().Errorf(ctx, "InitPassword: Failed to update password for UserID %d: %v", userId, err)
		return gerror.Wrapf(err, "Failed to update password for UserID %d", userId)
	}

	g.Log().Infof(ctx, "Password for UserID %d has been reset by an admin.", userId)

	// Optionally: Invalidate user's existing tokens/sessions here by calling TokenService.KickUserTokens(ctx, userId)
	// tokenService := NewTokenService()
	// if errKick := tokenService.KickUserTokens(ctx, userId); errKick != nil {
	// 	g.Log().Warningf(ctx, "InitPassword: Failed to kick existing tokens for UserID %d after password reset: %v", userId, errKick)
		// Decide if this should be a critical error. For now, just log.
	// }

	return nil
}

// ModifyPassword allows an authenticated user to change their own password.
func (s *UserService) ModifyPassword(ctx context.Context, userId uint, oldPassword string, newPassword string) error {
	// 1. Basic validation for inputs
	if userId == 0 { // Should be caught by middleware, but good practice
		return gerror.NewCodef(gcode.CodeInvalidArgument, g.I18n().T(ctx, "user.idInvalid"))
	}
	if oldPassword == "" {
		return gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.oldPasswordRequired"))
	}
	if len(newPassword) < 6 || len(newPassword) > 30 { // Assuming same constraints
		return gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "validation.passwordLength"))
	}
	if oldPassword == newPassword {
		return gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.newPasswordSameAsOld"))
	}


	// 2. Fetch the user
	user, err := dao.User.GetById(ctx, userId)
	if err != nil {
		return gerror.Wrapf(err, "ModifyPassword: Failed to retrieve user with ID %d", userId)
	}
	if user == nil {
		return gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "user.notFoundId", userId))
	}

	// 3. Verify oldPassword against the user's stored hashed password
	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(oldPassword))
	if err != nil { // Handles mismatch (ErrMismatchedHashAndPassword)
		return gerror.NewCode(gcode.CodeValidationFailed, g.I18n().T(ctx, "auth.oldPasswordIncorrect"))
	}

	// 4. Hash the newPassword
	hashedNewPassword, bcryptErr := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if bcryptErr != nil {
		g.Log().Errorf(ctx, "ModifyPassword: Failed to hash new password for UserID %d: %v", userId, bcryptErr)
		return gerror.Wrap(bcryptErr, "Failed to hash new password")
	}

	// 5. Update the user's password in the database
	updateData := g.Map{
		dao.User.Columns.Password:  string(hashedNewPassword),
		dao.User.Columns.UpdatedAt: gtime.Now(),
	}
	_, err = dao.User.Ctx(ctx).Data(updateData).Where(dao.User.Columns.Id, userId).Update()
	if err != nil {
		g.Log().Errorf(ctx, "ModifyPassword: Failed to update password for UserID %d: %v", userId, err)
		return gerror.Wrapf(err, "Failed to update password for UserID %d", userId)
	}

	g.Log().Infof(ctx, "Password for UserID %d has been changed by the user.", userId)

	// 6. Invalidate user's existing tokens/sessions
	tokenService := NewTokenService() // Assuming NewTokenService() returns ITokenService from internal/service
	if errKick := tokenService.KickUserTokens(ctx, userId); errKick != nil {
		g.Log().Warningf(ctx, "ModifyPassword: Failed to kick existing tokens for UserID %d after password change: %v", userId, errKick)
		// Not returning this error as primary operation (password change) was successful.
		// This could be logged to an audit trail or monitoring system.
	} else {
		g.Log().Infof(ctx, "ModifyPassword: Successfully kicked existing tokens for UserID %d.", userId)
	}

	return nil
}


// Note: Ensure `user_dao.go` has methods like GetByUsername, GetByEmailExcludeId, List, Count, Create, Update, Delete.
// Note: Ensure `NewTokenService()` is correctly referencing the token service from `internal/service/token_service.go`.
// Note: Ensure `Casbin()` is correctly referencing the casbin service from `internal/app/service/casbin_service.go`.
// Need to import "github.com/gogf/gf/v2/util/gconv" for gconv.String().
