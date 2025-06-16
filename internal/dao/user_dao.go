package dao

import (
	"context"
	"database/sql"
	"fmt"          // Added for fmt.Errorf in Create if LastInsertId fails
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/errors/gerror"
)

var _ gdb.DB = g.DB() // Prevent "imported and not used" for gdb if only used in Model()

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
		if gerror.Is(err, sql.ErrNoRows) {
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
		if gerror.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an application error
		}
		return nil, err // Other DB errors
	}
	return user, nil // User found, no error
}

// Update updates selected fields of a user record.
func (d *UserDao) Update(ctx context.Context, user *entity.User) error {
	updateData := g.Map{
		"nickname": user.Nickname,
		"email":    user.Email,
		"status":   user.Status,
	}
	// If password needs to be updatable, it should be handled separately
	// and include hashing logic, not directly setting PasswordHash from entity.
	// if user.PasswordHash != "" {
	//    updateData["password_hash"] = user.PasswordHash
	// }
	_, err := g.DB().Model("users").
		Data(updateData).
		Where("id", user.Id).
		Update()
	return err
}

// Delete hard deletes a user by ID.
func (d *UserDao) Delete(ctx context.Context, id uint64) error {
	_, err := g.DB().Model("users").Where("id", id).Delete()
	return err
}

// List retrieves a list of users with pagination.
func (d *UserDao) List(ctx context.Context, in *input.UserListInput) (users []*entity.User, total int, err error) {
	query := g.DB().Model("users")

	total, err = query.Count()
	if err != nil {
		return nil, 0, gerror.Wrap(err, "UserDao: Failed to count users")
	}
	if total == 0 {
		return []*entity.User{}, 0, nil
	}

	err = query.Page(in.Page, in.PageSize).Order("id ASC").Scan(&users)
	if err != nil {
		return nil, 0, gerror.Wrap(err, "UserDao: Failed to list users")
	}
	return users, total, nil
}
