package user

import (
	"context"
	// "database/sql" // No longer directly needed here as DAO handles ErrNoRows
	"fmt"
	"golang.org/x/crypto/bcrypt"
	// "github.com/gogf/gf/v2/database/gdb" // No longer explicitly needed if using sql.ErrNoRows via DAO and gerror.Is
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"yuncms/internal/dao"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"yuncms/internal/model/output"
    "yuncms/internal/utility/i18nutil"
)

var _ = bcrypt.DefaultCost
var _ = fmt.Errorf("")

type UserLogic struct {
	userDao *dao.UserDao
}

func NewUserLogic() *UserLogic {
	return &UserLogic{userDao: dao.NewUserDao()}
}

// CreateUser creates a new user.
func (ul *UserLogic) CreateUser(ctx context.Context, in *input.UserCreateInput) (*entity.User, error) {
	existingUser, err := ul.userDao.GetByUsername(ctx, in.Username) // DAO returns (nil, nil) on not found
	if err != nil { // This implies a DB error other than "not found"
		return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db_check_username", "Username", in.Username))
	}
	if existingUser != nil {
		return nil, gerror.New(i18nutil.T(ctx, "error_username_exists", "Username", in.Username))
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_password_hash_failed"))
	}

	user := &entity.User{
		Username:     in.Username,
		PasswordHash: string(hashedPassword),
		Nickname:     in.Nickname,
		Email:        in.Email,
		Status:       in.Status,
	}

	err = ul.userDao.Create(ctx, user)
	if err != nil {
		return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db_create_user"))
	}

	g.Log().Debugf(ctx, "User created in logic with ID: %d", user.Id)
	user.PasswordHash = ""
	return user, nil
}

// GetUserByUsername retrieves a user by username.
func (ul *UserLogic) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
    user, err := ul.userDao.GetByUsername(ctx, username) // DAO returns (nil, nil) on not found
    if err != nil { // This implies a DB error other than "not found"
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db_get_user_by_username", "Username", username))
    }
    if user == nil { // This implies "not found"
        return nil, gerror.New(i18nutil.T(ctx, "error_user_not_found"))
    }
    user.PasswordHash = ""
    return user, nil
}

// GetUserById retrieves a user by ID.
func (ul *UserLogic) GetUserById(ctx context.Context, id uint64) (*entity.User, error) {
    user, err := ul.userDao.GetById(ctx, id) // DAO returns (nil, nil) on not found
    if err != nil { // This means a DB error other than "not found" occurred
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
    }
    if user == nil { // This means "not found" (DAO returned nil, nil)
         return nil, gerror.New(i18nutil.T(ctx, "error_user_not_found"))
    }
    user.PasswordHash = ""
    return user, nil
}

// UpdateUser updates an existing user.
func (ul *UserLogic) UpdateUser(ctx context.Context, id uint64, in *input.UserUpdateInput) error {
	existingUser, err := ul.userDao.GetById(ctx, id) // DAO returns (nil, nil) on not found
	if err != nil { // This means a DB error other than "not found" occurred
		return gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
	}
    if existingUser == nil { // This means "not found"
         return gerror.New(i18nutil.T(ctx, "error_user_not_found"))
    }

	if in.Nickname != nil {
		existingUser.Nickname = *in.Nickname
	}
	if in.Email != nil {
		existingUser.Email = *in.Email
	}
	if in.Status != nil {
		existingUser.Status = *in.Status
	}

	err = ul.userDao.Update(ctx, existingUser)
	if err != nil {
		return gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
	}
	return nil
}

// DeleteUser deletes a user by ID.
func (ul *UserLogic) DeleteUser(ctx context.Context, id uint64) error {
	existingUser, err := ul.userDao.GetById(ctx, id) // Check existence first
    if err != nil { // This means a DB error other than "not found" occurred
        return gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
    }
    if existingUser == nil { // This means "not found"
        return gerror.New(i18nutil.T(ctx, "error_user_not_found"))
    }

	err = ul.userDao.Delete(ctx, id)
	if err != nil {
		return gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
	}
	return nil
}

// ListUsers retrieves a list of users with pagination.
func (ul *UserLogic) ListUsers(ctx context.Context, in *input.UserListInput) (*output.UserListOutput, error) {
	users, total, err := ul.userDao.List(ctx, in)
	if err != nil {
		return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
	}
    for _, u := range users {
        if u != nil {
            u.PasswordHash = "" // Clear sensitive data
        }
    }
	return &output.UserListOutput{
		List:  users,
		Total: total,
		Page:  in.Page,
		Size:  in.PageSize,
	}, nil
}
