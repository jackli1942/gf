package internal

import (
	"context"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"yuncms/internal/model/entity" // Assuming entity path
)

// DepartmentDao is data access object for table department data operations.
type DepartmentDao struct {
	table   string          // table is the underlying table name of the DAO.
	group   string          // group is the database configuration group name of current DAO.
	columns DepartmentColumns // columns contains all the column names of Table for convenient usage.
}

// DepartmentColumns defines and stores column names for table department.
type DepartmentColumns struct {
	Id        string // Department ID
	ParentId  string // Parent Department ID (0 for root)
	Name      string // Department Name
	Sort      string // Sort order; smaller is higher priority
	Status    string // Status (1:active, 0:disabled)
	Remark    string // Optional remarks for the department
	CreatedAt string // Creation time
	UpdatedAt string // Last update time
}

// NewDepartmentDao creates and returns a new DAO object for table data access.
func NewDepartmentDao() *DepartmentDao {
	return &DepartmentDao{
		group:   "default",                     // Default database group
		table:   "department",                  // Table name
		columns: departmentColumns,             // Column names
	}
}

// DB retrieves and returns the underlying raw database management object of current DAO.
func (dao *DepartmentDao) DB() gdb.DB {
	return g.DB(dao.group)
}

// Table returns the table name of current DAO.
func (dao *DepartmentDao) Table() string {
	return dao.table
}

// Columns returns all column names of current DAO.
func (dao *DepartmentDao) Columns() DepartmentColumns {
	return dao.columns
}

// Group returns the configuration group name of database of current DAO.
func (dao *DepartmentDao) Group() string {
	return dao.group
}

// Ctx creates and returns the Model for current DAO, It automatically sets the context for current operation.
func (dao *DepartmentDao) Ctx(ctx context.Context) *gdb.Model {
	return dao.DB().Model(dao.table).Safe().Ctx(ctx)
}

// Transaction wraps the transaction logic using function f.
// It rollbacks the transaction and returns the error from function f if it returns non-nil error.
// It commits the transaction and returns nil if function f returns nil.
//
// Note that, you should not Commit or Rollback the transaction in function f
// as it is automatically handled by this function.
func (dao *DepartmentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return dao.Ctx(ctx).Transaction(ctx, f)
}

// FillRecord fills an entity object with current record from the given gdb.Record.
// This is used for custom DAO logic.
func (dao *DepartmentDao) FillRecord(r *gdb.Record, e *entity.Department) {
	*e = entity.Department{}
	_ = r.Struct(e)
}

// departmentColumns holds the column names mapping.
var departmentColumns = DepartmentColumns{
	Id:        "id",
	ParentId:  "parent_id",
	Name:      "name",
	Sort:      "sort",
	Status:    "status",
	Remark:    "remark",
	CreatedAt: "created_at",
	UpdatedAt: "updated_at",
}
```
