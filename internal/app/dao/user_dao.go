package dao

import (
	"context"
	"yuncms/internal/app/model" // Path to user_entity.go's package

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gcode" // Added import
	"github.com/gogf/gf/v2/errors/gerror" // Added import
	"github.com/gogf/gf/v2/frame/g"
)

type userColumns struct {
	Id        string
	Uuid      string
	Username  string
	Password  string
	Nickname  string
	Email     string
	Phone     string
	Avatar    string
	Status    string
	CreatedAt string
	UpdatedAt string
}

type UserDao struct {
	table   string
	Columns userColumns
}

var (
	// User is the singleton instance of the UserDao.
	// The table name for users is "users".
	User = NewUserDao("users")
)

// NewUserDao creates and returns a new DAO object for user operations.
func NewUserDao(table string) *UserDao {
	return &UserDao{
		table: table,
		Columns: userColumns{
			Id:        "id",
			Uuid:      "uuid",
			Username:  "username",
			Password:  "password",
			Nickname:  "nickname",
			Email:     "email",
			Phone:     "phone",
			Avatar:    "avatar",
			Status:    "status",
			CreatedAt: "created_at",
			UpdatedAt: "updated_at",
		},
	}
}

// M returns a model instance for the DAO's table with context.
// It provides safe database operations.
func (d *UserDao) M(ctx context.Context) *gdb.Model {
	return g.DB().Model(d.table).Safe().Ctx(ctx)
}

// Create inserts a new user record into the database.
// It returns the ID of the newly inserted record and an error if any.
// This is a placeholder and needs to be implemented.
func (d *UserDao) Create(ctx context.Context, data *model.User) (id int64, err error) {
	lastInsertId, err := d.M(ctx).Data(data).InsertAndGetId()
	if err != nil {
		return 0, err
	}
	return lastInsertId, nil
}

// GetById retrieves a user by their ID.
func (d *UserDao) GetById(ctx context.Context, id uint) (user *model.User, err error) {
	err = d.M(ctx).WherePri(id).Scan(&user)
	return user, err
}

// GetByUsername retrieves a user by their username.
func (d *UserDao) GetByUsername(ctx context.Context, username string) (user *model.User, err error) {
	err = d.M(ctx).Where(d.Columns.Username, username).Scan(&user)
	return user, err
}

// GetByUuid retrieves a user by their UUID.
func (d *UserDao) GetByUuid(ctx context.Context, uuid string) (user *model.User, err error) {
	err = d.M(ctx).Where(d.Columns.Uuid, uuid).Scan(&user)
	return user, err
}

// Count retrieves the total number of users.
func (d *UserDao) Count(ctx context.Context) (total int, err error) {
	total, err = d.M(ctx).Count()
	return total, err
}

// List retrieves a list of users with pagination.
func (d *UserDao) List(ctx context.Context, offset int, limit int) (users []*model.User, err error) {
	err = d.M(ctx).Offset(offset).Limit(limit).Scan(&users)
	return users, err
}

// Update modifies an existing user record by ID.
// data should be a g.Map containing fields to update.
func (d *UserDao) Update(ctx context.Context, id uint, data g.Map) (err error) {
	_, err = d.M(ctx).Data(data).WherePri(id).Update()
	return err
}

// GetByEmailExcludeId retrieves a user by email, excluding a specific user ID.
func (d *UserDao) GetByEmailExcludeId(ctx context.Context, email string, excludeId uint) (user *model.User, err error) {
	err = d.M(ctx).Where(d.Columns.Email, email).WhereNot(d.Columns.Id, excludeId).Scan(&user)
	return user, err
}

// GetByPhoneExcludeId retrieves a user by phone, excluding a specific user ID.
// Assumes phone numbers should be unique if this method is used by the service.
func (d *UserDao) GetByPhoneExcludeId(ctx context.Context, phone string, excludeId uint) (user *model.User, err error) {
	err = d.M(ctx).Where(d.Columns.Phone, phone).WhereNot(d.Columns.Id, excludeId).Scan(&user)
	return user, err
}

// Delete removes a user by their ID.
// It returns gcode.NotFound if no rows were affected.
func (d *UserDao) Delete(ctx context.Context, id uint) (err error) {
	result, err := d.M(ctx).WherePri(id).Delete()
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		// This error is for RowsAffected() itself, not the delete operation
		return gerror.Wrap(err, "Failed to check rows affected after delete")
	}
	if rowsAffected == 0 {
		return gerror.NewCodef(gcode.CodeNotFound, "User with ID %d not found for deletion", id) // Ensure CodeNotFound
	}
	return nil
}
