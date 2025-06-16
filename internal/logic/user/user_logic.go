package user

import (
	"context"
	// "fmt" // Not used
	"yuncms/internal/dao"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"golang.org/x/crypto/bcrypt"
	// "github.com/gogf/gf/v2/database/gdb" // No longer directly needed here
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g" // Ensure g is imported for g.Log()
)

type UserLogic struct {
	userDao *dao.UserDao
}

func NewUserLogic() *UserLogic {
	return &UserLogic{userDao: dao.NewUserDao()}
}

// CreateUser creates a new user.
func (ul *UserLogic) CreateUser(ctx context.Context, in *input.UserCreateInput) (*entity.User, error) {
	existingUser, err := ul.userDao.GetByUsername(ctx, in.Username)
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to check existing username during DB query")
	}
	if existingUser != nil {
		return nil, gerror.New("Username already exists")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(in.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to hash password")
	}

	user := &entity.User{
		Username:     in.Username,
		PasswordHash: string(hashedPassword),
		Nickname:     in.Nickname,
		Email:        in.Email,
		Status:       in.Status,
	}

	err = ul.userDao.Create(ctx, user) // user.Id will be populated by DAO method
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to create user in DB")
	}

	g.Log().Debugf(ctx, "User created in logic with ID: %d", user.Id)
	user.PasswordHash = "" // Clear password hash before returning
	return user, nil
}

// GetUserByUsername retrieves a user by username.
func (ul *UserLogic) GetUserByUsername(ctx context.Context, username string) (*entity.User, error) {
    user, err := ul.userDao.GetByUsername(ctx, username)
    if err != nil {
        return nil, gerror.Wrapf(err, "Failed to get user by username '%s'", username)
    }
    if user == nil {
        return nil, gerror.Newf("User '%s' not found", username)
    }
    return user, nil
}
