package dao

import (
	"yuncms/internal/dao/internal" // Import the internal DAO
	"yuncms/internal/model/do" // Import DO for potential custom methods if any
	"yuncms/internal/model/entity" // Import Entity for potential custom methods if any

	// "context" // Example: if you add custom methods needing context
	// "github.com/gogf/gf/v2/database/gdb" // Example: if you add custom methods needing gdb
	// "github.com/gogf/gf/v2/frame/g"      // Example: if you add custom methods needing g.Map
)

// departmentDao is the data access object for department data operations.
// You can define custom methods here that are not auto-generated.
type departmentDao struct {
	*internal.DepartmentDao // Embed internal DAO for basic functionalities
}

var (
	// Department is the singleton instance of departmentDao.
	Department = departmentDao{
		internal.NewDepartmentDao(), // Initialize with internal DAO
	}
)

// You can add custom DAO methods here if needed. For example:
/*
// GetActiveDepartments retrieves all active departments.
func (d *departmentDao) GetActiveDepartments(ctx context.Context) ([]*entity.Department, error) {
	var departments []*entity.Department
	err := d.Ctx(ctx).Where(d.Columns().Status, 1).OrderAsc(d.Columns().Sort).Scan(&departments)
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to retrieve active departments")
	}
	return departments, nil
}

// GetDepartmentChildren retrieves direct children of a department.
func (d *departmentDao) GetDepartmentChildren(ctx context.Context, parentId uint) ([]*entity.Department, error) {
	var children []*entity.Department
	err := d.Ctx(ctx).Where(d.Columns().ParentId, parentId).OrderAsc(d.Columns().Sort).Scan(&children)
	if err != nil {
		return nil, gerror.Wrapf(err, "Failed to retrieve children for department ID %d", parentId)
	}
	return children, nil
}

// UpdateDepartmentName updates the name of a specific department.
// This is just an example of a custom method that might use a DO.
func (d *departmentDao) UpdateDepartmentName(ctx context.Context, departmentId uint, newName string) error {
	_, err := d.Ctx(ctx).Data(do.Department{Name: newName}).Where(d.Columns().Id, departmentId).Update()
	if err != nil {
		return gerror.Wrapf(err, "Failed to update name for department ID %d", departmentId)
	}
	return nil
}
*/
```
