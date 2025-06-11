package logic

import (
	"context"
	"errors" // For errors.Is
	"fmt"
	"sort"
	"strings"

	"gf_project/api/v1/system" // For DeptTreeItem struct
	"gf_project/internal/modules/system/model/do"
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"gf_project/internal/modules/system/model/internal/dao"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
	"github.com/gogf/gf/v2/util/gconv"
)

type sSystemDept struct{}

var DeptService = sSystemDept{}

// buildDeptTree recursively builds a tree structure from a list of departments.
// It also populates Value and Label for frontend tree components.
func buildDeptTree(ctx context.Context, parentId uint64, allDepts []*entity.SystemDept) []*system.DeptTreeItem {
	tree := make([]*system.DeptTreeItem, 0)
	for _, deptEntity := range allDepts {
		if deptEntity.ParentId == parentId {
			node := &system.DeptTreeItem{
				SystemDept: &entity.SystemDept{}, // Initialize embedded struct
				Value:      deptEntity.Id,
				Label:      deptEntity.Name,
			}
			_ = gconv.Struct(deptEntity, node.SystemDept) // Copy fields

			childChildren := buildDeptTree(ctx, deptEntity.Id, allDepts)
			if len(childChildren) > 0 {
				node.Children = childChildren
			}
			tree = append(tree, node)
		}
	}
	sort.SliceStable(tree, func(i, j int) bool {
		if tree[i].Sort != tree[j].Sort {
			return tree[i].Sort < tree[j].Sort
		}
		return tree[i].Id < tree[j].Id
	})
	return tree
}

// getAncestorsPath calculates the ancestor path string for a department.
// Devinggo's 'level' field stores this path, e.g., ",grandparent_id,parent_id,"
func (s *sSystemDept) getAncestorsPath(ctx context.Context, parentId uint64) (string, error) {
	if parentId == 0 {
		return ",0,", nil // Root department's conventional ancestor path including a common root "0"
	}
	parentDept, err := s.GetDeptById(ctx, parentId) // Uses the service method which handles "not found"
	if err != nil {
		return "", gerror.Wrapf(err, "parent department with ID %d not found for ancestor calculation", parentId)
	}
	// Parent's Level (ancestor path) + ParentID + ","
	return parentDept.Level + gconv.String(parentId) + ",", nil
}

func (s *sSystemDept) CreateDept(ctx context.Context, in *input.DeptCreateInp) (deptId uint64, err error) {
	count, err := dao.SystemDept.Ctx(ctx).Where(do.SystemDept{ParentId: in.ParentId, Name: in.Name}).Count()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to check department name existence")
	}
	if count > 0 {
		return 0, gerror.Newf("department name '%s' already exists under the selected parent", in.Name)
	}

	ancestorsPath, err := s.getAncestorsPath(ctx, in.ParentId)
	if err != nil {
		return 0, err
	}

	// Status: input struct uses 1 for Normal, 2 for Disabled. DB schema in devinggo also uses 1/2.
	// Our standardized input for User used 0/1. For Dept, we follow devinggo's 1/2 from input.
	newDept := do.SystemDept{
		ParentId: in.ParentId, Name: in.Name, Leader: in.Leader, Phone: in.Phone,
		Status: in.Status, Sort: in.Sort, Remark: in.Remark,
		Level:     ancestorsPath, // This is the ancestor path
		CreatedAt: gtime.Now(), UpdatedAt: gtime.Now(),
	}

	result, err := dao.SystemDept.Ctx(ctx).Data(newDept).Insert()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to create department")
	}

	newDeptId64, err := result.LastInsertId()
	if err != nil {
		return 0, gerror.Wrap(err, "failed to retrieve last insert ID for department")
	}
	return uint64(newDeptId64), nil
}

func (s *sSystemDept) UpdateDept(ctx context.Context, in *input.DeptUpdateInp) (err error) {
	dept, err := s.GetDeptById(ctx, in.Id)
	if err != nil {
		return err
	}

	if in.Name != dept.Name || in.ParentId != dept.ParentId { // Check name uniqueness only if name or parent changed
		countName, errName := dao.SystemDept.Ctx(ctx).Where(do.SystemDept{ParentId: in.ParentId, Name: in.Name}).AndNot(dao.SystemDept.Columns().Id, in.Id).Count()
		if errName != nil {
			return gerror.Wrap(errName, "failed to check department name uniqueness on update")
		}
		if countName > 0 {
			return gerror.Newf("department name '%s' already exists under parent ID %d", in.Name, in.ParentId)
		}
	}

	newAncestorsPath := dept.Level
	parentChanged := in.ParentId != dept.ParentId

	if parentChanged {
		if in.ParentId == dept.Id {
			return gerror.New("cannot move a department under itself")
		}

		parentAncestorsPath, calcErr := s.getAncestorsPath(ctx, in.ParentId)
		if calcErr != nil {
			return gerror.Wrap(calcErr, "failed to calculate new ancestors path for department update")
		}
		newAncestorsPath = parentAncestorsPath

		// Check if new parent is a descendant of the current department
		// Current department's path to its children would start with its own old ancestor path + its ID + ","
		// Example: if dept ID is 3, and its old level was ",0,1,", then its children's path starts with ",0,1,3,"
		pathPrefixOfCurrentDeptChildren := dept.Level + gconv.String(dept.Id) + ","
		if strings.HasPrefix(newAncestorsPath, pathPrefixOfCurrentDeptChildren) {
			return gerror.New("cannot move a department under one of its own children")
		}
	}

	updateData := do.SystemDept{
		Id: in.Id, ParentId: in.ParentId, Name: in.Name, Leader: in.Leader,
		Phone: in.Phone, Status: in.Status, Sort: in.Sort, Remark: in.Remark,
		Level: newAncestorsPath, UpdatedAt: gtime.Now(),
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, txErr := dao.SystemDept.Ctx(ctx).TX(tx).Data(updateData).Where(dao.SystemDept.Columns().Id, in.Id).Update()
		if txErr != nil {
			return txErr
		}

		if parentChanged {
			oldPathPrefixForChildren := dept.Level + gconv.String(dept.Id) + ","
			newPathPrefixForChildren := newAncestorsPath + gconv.String(dept.Id) + ","

			var childrenToUpdate []*entity.SystemDept
			errScan := dao.SystemDept.Ctx(ctx).TX(tx).WhereLike(dao.SystemDept.Columns().Level, oldPathPrefixForChildren+"%").Scan(&childrenToUpdate)
			if errScan != nil {
				return gerror.Wrap(errScan, "failed to retrieve children for ancestor update")
			}

			for _, child := range childrenToUpdate {
				updatedChildLevel := strings.Replace(child.Level, oldPathPrefixForChildren, newPathPrefixForChildren, 1)
				_, updateChildErr := dao.SystemDept.Ctx(ctx).TX(tx).
					Where(dao.SystemDept.Columns().Id, child.Id).
					Data(do.SystemDept{Level: updatedChildLevel, UpdatedAt: gtime.Now()}).
					Update()
				if updateChildErr != nil {
					return gerror.Wrapf(updateChildErr, "failed to update Level for child department ID %d", child.Id)
				}
			}
		}
		return nil
	})
}

func (s *sSystemDept) DeleteDept(ctx context.Context, id uint64) (err error) {
	dept, err := s.GetDeptById(ctx, id)
	if err != nil {
		return err
	}

	// Path segment representing this department and its potential children
	// e.g., if dept.Id is 2, and dept.Level is ",0,1,", then path segment is ",0,1,2,"
	pathSegmentOfThisDeptAndChildren := dept.Level + gconv.String(id) + ","

	var allDeptsInSubtree []*entity.SystemDept
	err = dao.SystemDept.Ctx(ctx).WhereLike(dao.SystemDept.Columns().Level, pathSegmentOfThisDeptAndChildren+"%").OrWhere(dao.SystemDept.Columns().Id, id).Scan(&allDeptsInSubtree)
	if err != nil {
		return gerror.Wrap(err, "failed to fetch department subtree for user check")
	}

	var deptIdsInSubtree []uint64
	if len(allDeptsInSubtree) == 0 { // Should at least contain the dept itself if found
		deptIdsInSubtree = []uint64{id}
	} else {
		for _, d := range allDeptsInSubtree {
			deptIdsInSubtree = append(deptIdsInSubtree, d.Id)
		}
	}

	userCount, err := dao.SystemUser.Ctx(ctx).WhereIn(dao.SystemUser.Columns().DeptId, deptIdsInSubtree).Count()
	if err != nil {
		return gerror.Wrap(err, "failed to check for users in department before deletion")
	}
	if userCount > 0 {
		return gerror.Newf("cannot delete department: %d users are assigned to this department or its sub-departments", userCount)
	}

	return g.DB().Transaction(ctx, func(ctx context.Context, tx gdb.TX) error {
		_, txErr := dao.SystemDept.Ctx(ctx).TX(tx).WhereIn(dao.SystemDept.Columns().Id, deptIdsInSubtree).Delete()
		if txErr != nil {
			return gerror.Wrap(txErr, "failed to delete department and its children from DB")
		}
		return nil
	})
}

func (s *sSystemDept) GetDeptById(ctx context.Context, id uint64) (dept *entity.SystemDept, err error) {
	if id == 0 { // Convention for "root" or "no parent"
		return &entity.SystemDept{Id: 0, Name: "Root", Level: ","}, nil // Return a virtual root node
	}
	err = dao.SystemDept.Ctx(ctx).Where(dao.SystemDept.Columns().Id, id).Scan(&dept)
	if err != nil {
		if errors.Is(err, gdb.ErrNoRows) {
			return nil, gerror.Newf("department with ID %d not found", id)
		}
		return nil, gerror.Wrapf(err, "failed to retrieve department by ID %d", id)
	}
	return dept, nil
}

func (s *sSystemDept) getAllDepts(ctx context.Context, filter *input.DeptListInp) (list []*entity.SystemDept, err error) {
	m := dao.SystemDept.Ctx(ctx).OrderAsc(dao.SystemDept.Columns().Sort).OrderAsc(dao.SystemDept.Columns().Id)
	if filter != nil {
		if filter.Name != "" {
			m = m.WhereLike(dao.SystemDept.Columns().Name, "%"+filter.Name+"%")
		}
		if filter.Status != nil {
			m = m.Where(dao.SystemDept.Columns().Status, *filter.Status)
		} // Assumes Status in input is 0/1, DB is 1/2
	}
	err = m.Scan(&list)
	if err != nil {
		return nil, gerror.Wrap(err, "failed to retrieve all department items")
	}
	return list, nil
}

func (s *sSystemDept) GetDeptList(ctx context.Context, in *input.DeptListInp) ([]*system.DeptTreeItem, error) {
	allDepts, err := s.getAllDepts(ctx, in)
	if err != nil {
		return nil, err
	}
	// The root for tree building is ParentId = 0 in our DB.
	// The virtual root node (ID 0, Level ",") from getAncestorsPath means actual top-level depts have ParentId 0.
	return buildDeptTree(ctx, 0, allDepts), nil
}
