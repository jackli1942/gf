package department

import (
	"context"
	"yuncms/internal/dao"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"yuncms/internal/model/output"
    "yuncms/internal/utility/i18nutil"
	"github.com/gogf/gf/v2/errors/gerror"
    // "database/sql" // No longer needed here, DAO handles sql.ErrNoRows
)

type DepartmentLogic struct {
	deptDao *dao.DepartmentDao
}

func NewDepartmentLogic() *DepartmentLogic {
	return &DepartmentLogic{deptDao: dao.NewDepartmentDao()}
}

func (dl *DepartmentLogic) Create(ctx context.Context, in *input.DepartmentCreateInput) (*entity.Department, error) {
	dept := &entity.Department{
		ParentId:  in.ParentId,
		Name:      in.Name,
		Leader:    in.Leader,
		Status:    in.Status,
		SortOrder: in.SortOrder,
	}
	id, err := dl.deptDao.Create(ctx, dept)
	if err != nil {
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db_create_department")) // New i18n key
    }
	dept.Id = uint64(id)
	return dept, nil
}

func (dl *DepartmentLogic) GetById(ctx context.Context, id uint64) (*output.DepartmentOutput, error) {
	dept, err := dl.deptDao.GetById(ctx, id)
	if err != nil { // This is a DB error other than "not found"
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
    }
	if dept == nil { // This means "not found"
        return nil, gerror.New(i18nutil.T(ctx, "department_not_found"))
    }
	return &output.DepartmentOutput{Department: dept}, nil
}

func (dl *DepartmentLogic) Update(ctx context.Context, id uint64, in *input.DepartmentUpdateInput) error {
    dept, err := dl.deptDao.GetById(ctx, id)
    if err != nil { return gerror.Wrap(err, i18nutil.T(ctx, "error_db")) }
    if dept == nil { return gerror.New(i18nutil.T(ctx, "department_not_found")) }

    if in.ParentId != nil { dept.ParentId = *in.ParentId }
    if in.Name != nil { dept.Name = *in.Name }
    if in.Leader != nil { dept.Leader = *in.Leader }
    if in.Status != nil { dept.Status = *in.Status }
    if in.SortOrder != nil { dept.SortOrder = *in.SortOrder }

    err = dl.deptDao.Update(ctx, dept)
    if err != nil {
        return gerror.Wrap(err, i18nutil.T(ctx, "error_db_update_department")) // New i18n key
    }
    return nil
}

func (dl *DepartmentLogic) Delete(ctx context.Context, id uint64) error {
    dept, err := dl.deptDao.GetById(ctx, id) // Check existence first
    if err != nil { return gerror.Wrap(err, i18nutil.T(ctx, "error_db")) }
    if dept == nil { return gerror.New(i18nutil.T(ctx, "department_not_found")) }

    // Optional: Check for child departments before deleting.
    // For now, direct delete. Add this logic if business rules require it.

    err = dl.deptDao.Delete(ctx, id)
    if err != nil {
        return gerror.Wrap(err, i18nutil.T(ctx, "error_db_delete_department")) // New i18n key
    }
    return nil
}

func (dl *DepartmentLogic) List(ctx context.Context, in *input.DepartmentListInput) (*output.DepartmentListOutput, error) {
    depts, total, err := dl.deptDao.List(ctx, in)
    if err != nil { return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db")) }

    var deptOutputs []*output.DepartmentOutput
    for _, d := range depts {
        deptOutputs = append(deptOutputs, &output.DepartmentOutput{Department: d})
    }

    return &output.DepartmentListOutput{
        List:  deptOutputs,
        Total: total,
        Page:  in.Page,
        Size:  in.PageSize,
    }, nil
}

// GetTree builds a tree structure from all departments.
// This is a more complex operation and might be refined or moved later.
func (dl *DepartmentLogic) GetTree(ctx context.Context) ([]*output.DepartmentOutput, error) {
    allDepts, err := dl.deptDao.GetAll(ctx)
    if err != nil {
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
    }
    return buildDeptTree(allDepts, 0), nil
}

// buildDeptTree is a helper function to recursively build the department tree.
func buildDeptTree(depts []*entity.Department, parentId uint64) []*output.DepartmentOutput {
    var tree []*output.DepartmentOutput
    for _, dept := range depts {
        if dept.ParentId == parentId {
            children := buildDeptTree(depts, dept.Id)
            node := &output.DepartmentOutput{
                Department: dept,
                Children:   children,
            }
            tree = append(tree, node)
        }
    }
    return tree
}
