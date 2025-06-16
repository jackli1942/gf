package dao

import (
	"context"
	"database/sql" // Added for sql.ErrNoRows
	"fmt"          // Added for fmt.Errorf
	"yuncms/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// Prevent "imported and not used" error for gdb if types are used implicitly.
var _ gdb.DB = g.DB()

type UserDao struct{}

func NewUserDao() *UserDao {
	return &UserDao{}
}

// Create creates a new user record and updates the user.Id with the new ID.
func (d *UserDao) Create(ctx context.Context, user *entity.User) error {
	res, err := g.DB().Model("users").Data(user).Insert()
	if err != nil {
		return err
	}
	newId, err := res.LastInsertId()
	if err != nil {
		return fmt.Errorf("failed to get last insert ID: %w", err)
	}
	user.Id = uint64(newId)
	return nil
}

// GetByUsername retrieves a user by username.
func (d *UserDao) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
	var user *entity.User
	err := g.DB().Model("users").Where("username", username).Scan(&user)
	if err != nil {
		if gerror.Is(err, sql.ErrNoRows) { // Changed to sql.ErrNoRows
			return nil, nil // Not found is not an application error
		}
		return nil, err // Other DB errors
	}
	return user, nil // User found, no error
}

// GetById retrieves a user by ID.
func (d *UserDao) GetById(ctx context.Context, id uint64) (*entity.User, error) {
	var user *entity.User
	err := g.DB().Model("users").Where("id", id).Scan(&user)
	if err != nil {
		if gerror.Is(err, sql.ErrNoRows) { // Changed to sql.ErrNoRows
			return nil, nil // Not found is not an application error
		}
		return nil, err // Other DB errors
	}
	return user, nil // User found, no error
}
