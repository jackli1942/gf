package dao

import (
	"context"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/errors/gerror"
    "database/sql"
    "fmt" // For errors in Create
)

var _ gdb.DB = g.DB()

type DepartmentDao struct{}

func NewDepartmentDao() *DepartmentDao { return &DepartmentDao{} }

func (d *DepartmentDao) Create(ctx context.Context, dept *entity.Department) (id int64, err error) {
	res, err := g.DB().Model("departments").Data(dept).Insert()
	if err != nil { return 0, err }
    newId, err := res.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to get last insert ID for department: %w", err)
    }
	return newId, nil
}

func (d *DepartmentDao) GetById(ctx context.Context, id uint64) (*entity.Department, error) {
	var dept *entity.Department
	err := g.DB().Model("departments").Where("id", id).Scan(&dept)
    if err != nil {
        if gerror.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
	return dept, nil
}

func (d *DepartmentDao) Update(ctx context.Context, dept *entity.Department) error {
	// Use Data map for selective updates, similar to UserDAO
    // This ensures zero values in 'dept' don't overwrite fields if not intended.
    // However, the logic layer will construct 'dept' with intended values.
    // If 'dept' itself is already only populated with desired changes, .Data(dept) is fine.
    // For this structure, assuming 'dept' has all fields correctly set for update.
	_, err := g.DB().Model("departments").Data(dept).Where("id", dept.Id).Update()
	return err
}

func (d *DepartmentDao) Delete(ctx context.Context, id uint64) error {
	_, err := g.DB().Model("departments").Where("id", id).Delete()
	return err
}

func (d *DepartmentDao) List(ctx context.Context, in *input.DepartmentListInput) (depts []*entity.Department, total int, err error) {
    query := g.DB().Model("departments")
    if in.Name != "" {
        query = query.Where("name LIKE ?", "%"+in.Name+"%")
    }
    if in.Status != 0 { // Assuming 0 means all, 1 active, 2 inactive
        query = query.Where("status", in.Status)
    }

    total, err = query.Count()
    if err != nil { return nil, 0, gerror.Wrap(err, "DepartmentDao: Count failed") }
    if total == 0 { return []*entity.Department{}, 0, nil }

    err = query.Page(in.Page, in.PageSize).Order("sort_order ASC, id ASC").Scan(&depts)
    if err != nil { return nil, 0, gerror.Wrap(err, "DepartmentDao: Scan failed") }
    return depts, total, nil
}

// GetAll retrieves all departments, typically for building a tree
func (d *DepartmentDao) GetAll(ctx context.Context) (depts []*entity.Department, err error) {
    err = g.DB().Model("departments").Order("sort_order ASC, id ASC").Scan(&depts)
    if err != nil { return nil, gerror.Wrap(err, "DepartmentDao: GetAll failed") }
    return depts, nil
}
