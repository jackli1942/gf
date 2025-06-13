package dao

import (
	"context"
	// "yuncms/internal/app/model/do"
	// "yuncms/internal/app/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
)

// DepartmentDao is the data access object for department operations.
type DepartmentDao struct {
	internalDepartmentDao // Embed the internal DAO
}

// internalDepartmentDao is the internal DAO for department.
type internalDepartmentDao struct {
	table   string
	group   string
	columns departmentColumns
}

// departmentColumns holds column names for the department table.
type departmentColumns struct {
	Id        string
	ParentId  string
	Name      string
	Sort      string
	Status    string
	Remark    string
	CreatedAt string
	UpdatedAt string
}

var (
	// Department is the singleton instance of DepartmentDao.
	Department = NewDepartmentDao()
)

// NewDepartmentDao creates and returns a new DAO object for department operations.
func NewDepartmentDao() *DepartmentDao {
	return &DepartmentDao{
		internalDepartmentDao: internalDepartmentDao{
			table: "department", // Table name placeholder
			group: "default",    // Database group placeholder
			columns: departmentColumns{
				Id:        "id",
				ParentId:  "parent_id",
				Name:      "name",
				Sort:      "sort",
				Status:    "status",
				Remark:    "remark",
				CreatedAt: "created_at",
				UpdatedAt: "updated_at",
			},
		},
	}
}

// Columns returns the column names for the department table.
func (d *DepartmentDao) Columns() departmentColumns {
	return d.columns
}

// Ctx returns a new model instance for the department table with context.
func (d *DepartmentDao) Ctx(ctx context.Context) *gdb.Model {
	return g.DB(d.group).Model(d.table).Safe().Ctx(ctx)
}

// Transaction executes the given function f within a database transaction.
func (d *DepartmentDao) Transaction(ctx context.Context, f func(ctx context.Context, tx gdb.TX) error) (err error) {
	return d.Ctx(ctx).Transaction(ctx, f)
}

// Fill with nil for tools if required.
func (d *DepartmentDao) Fill(data interface{}) error {
	return g.Struct(data, d)
}

// Clone creates a new DAO object for department operations.
func (d *DepartmentDao) Clone() *DepartmentDao {
	return &DepartmentDao{
		internalDepartmentDao: d.internalDepartmentDao,
	}
}
