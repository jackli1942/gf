package service

import (
	"context"
	"yuncms/internal/app/dao"
	"yuncms/internal/app/model"

	"github.com/gogf/gf/v2/errors/gcode" // Added import for gcode
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

// UserCreateInput is the DTO for creating a new user.
type UserCreateInput struct {
	Username string `json:"username" v:"required|length:3,30#Username is required|Username length must be between 3 and 30"`
	Password string `json:"password" v:"required|length:6,30#Password is required|Password length must be between 6 and 30"`
	Nickname string `json:"nickname" v:"required|length:1,30#Nickname is required"`
	Email    string `json:"email"    v:"email#Invalid email format"`
	Phone    string `json:"phone"`
	Avatar   string `json:"avatar"`
	Status   *int   `json:"status"   v:"in:0,1#Invalid status"` // Pointer to distinguish not-set from 0
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
	Nickname  *string `json:"nickname" v:"length:1,30#nickname_length_error"`
	Password  *string `json:"password" v:"length:6,30#password_length_error"`
	Email     *string `json:"email"    v:"email#email_format_error"`
	Phone     *string `json:"phone"    v:"phone-loose#phone_format_error"`
	Avatar    *string `json:"avatar"   v:"url#avatar_url_error"`
	Status    *int    `json:"status"   v:"in:0,1#status_invalid_error"`
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
	if err := g.Validator().Data(in).Run(ctx); err != nil {
		return nil, err // err is already gerror/gvalid.Error
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
		return nil, gerror.NewCodef(gcode.CodeNotFound, "Updated user with ID %d not found after update", id)
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
