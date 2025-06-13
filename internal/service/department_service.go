package service

import (
	"context"
	"sort"
	"yuncms/internal/dao"
	"yuncms/internal/model/entity"

	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
)

// DepartmentService manages department related business logic.
type DepartmentService struct{}

// NewDepartmentService creates and returns a new DepartmentService.
func NewDepartmentService() *DepartmentService {
	return &DepartmentService{}
}

// DepartmentListInput defines the input for listing departments.
type DepartmentListInput struct {
	Name   string `json:"name,omitempty" description:"Filter by department name (fuzzy match)"`
	Status *int   `json:"status,omitempty" v:"in:0,1#department.statusInvalid" description:"Filter by status (0:disabled, 1:active)"`
}

// DepartmentListItem defines the structure for a department item in a list, including children for tree view.
type DepartmentListItem struct {
	entity.Department                 // Embed the Department entity
	Children          []*DepartmentListItem `json:"children,omitempty"`
}

// ListDepartments retrieves a list of departments, optionally filtered, and returns them as a tree structure.
func (s *DepartmentService) ListDepartments(ctx context.Context, in DepartmentListInput) (list []*DepartmentListItem, err error) {
	// 1. Construct query
	m := dao.Department.Ctx(ctx)

	if in.Name != "" {
		m = m.WhereLike(dao.Department.Columns().Name, "%"+in.Name+"%")
	}
	if in.Status != nil {
		m = m.Where(dao.Department.Columns().Status, *in.Status)
	}

	// 2. Order and retrieve all matching departments
	var allDepts []*entity.Department
	err = m.OrderAsc(dao.Department.Columns().Sort).OrderAsc(dao.Department.Columns().Id).Scan(&allDepts)
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to retrieve departments from database")
	}

	if len(allDepts) == 0 {
		return []*DepartmentListItem{}, nil // Return empty slice if no departments found
	}

	// 3. Transform flat list to tree structure
	deptMap := make(map[uint]*DepartmentListItem)
	list = make([]*DepartmentListItem, 0)

	// First pass: create DepartmentListItem for each department and populate the map
	for _, deptEntity := range allDepts {
		deptMap[deptEntity.Id] = &DepartmentListItem{
			Department: *deptEntity,
			Children:   []*DepartmentListItem{}, // Initialize children slice
		}
	}

	// Second pass: build the tree structure
	for _, deptEntity := range allDepts {
		item := deptMap[deptEntity.Id] // Get the item from map (already created)
		if item.ParentId != 0 {
			if parent, ok := deptMap[item.ParentId]; ok {
				parent.Children = append(parent.Children, item)
			} else {
				// Parent not found in the current filtered list, treat as a root node for this list
				list = append(list, item)
			}
		} else {
			// It's a root node (ParentId is 0)
			list = append(list, item)
		}
	}

	// Sort children for consistent order if needed (already ordered by DB query globally)
	// If children need to be sorted explicitly within each parent after grouping:
	for _, item := range deptMap {
		if len(item.Children) > 1 {
			sort.Slice(item.Children, func(i, j int) bool {
				if item.Children[i].Sort == item.Children[j].Sort {
					return item.Children[i].Id < item.Children[j].Id
				}
				return item.Children[i].Sort < item.Children[j].Sort
			})
		}
	}


	return list, nil
}

// Helper function to ensure all DAO methods are available (conceptual)
func _() {
	// This function is just for type-checking during development
	// and ensures that the assumed DAO methods exist.
	var _ *gdb.Model = dao.Department.Ctx(nil)
	_ = dao.Department.Columns().Name
	_ = dao.Department.Columns().Status
	_ = dao.Department.Columns().Sort
	_ = dao.Department.Columns().Id
	var depts []*entity.Department
	_ = dao.Department.Ctx(nil).Scan(&depts)
}

// GetDepartmentById retrieves a single department by its ID.
func (s *DepartmentService) GetDepartmentById(ctx context.Context, id uint) (department *entity.Department, err error) {
	if id == 0 { // Basic validation, though controller should also validate
		return nil, gerror.NewCodef(gcode.CodeInvalidArgument, g.I18n().T(ctx, "department.idInvalid"))
	}

	err = dao.Department.Ctx(ctx).Where(dao.Department.Columns().Id, id).Scan(&department)
	if err != nil {
		// Log the actual DB error for debugging
		g.Log().Errorf(ctx, "GetDepartmentById: Failed to scan department for ID %d: %v", id, err)
		// Return a generic error or a not-found error if that's the case (sql.ErrNoRows)
		// GoFrame's Scan typically returns nil for 'department' and no error if no rows are found.
		// However, if there's a real DB error, err will be non-nil.
		return nil, gerror.Wrapf(err, "Database error while fetching department with ID %d", id)
	}
	if department == nil {
		// This means no rows were found and Scan successfully completed (err was nil).
		return nil, gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "department.notFoundId", id))
	}

	return department, nil
}

// DepartmentUpdateInput defines the input for updating an existing department.
type DepartmentUpdateInput struct {
	Name     *string `json:"name,omitempty" v:"required-without-all:ParentId,Sort,Status,Remark|length:1,100#department.nameRequiredWithoutAll|department.nameLength" description:"Department Name"`
	ParentId *uint   `json:"parentId,omitempty" v:"min:0#department.parentIdMin" description:"Parent Department ID (0 for root)"`
	Sort     *int    `json:"sort,omitempty" description:"Sort order"`
	Status   *int    `json:"status,omitempty" v:"in:0,1#department.statusInvalid" description:"Status (1:active, 0:disabled)"`
	Remark   *string `json:"remark,omitempty" v:"max-length:255#department.remarkMaxLength" description:"Optional remarks"`
}

// UpdateDepartment updates an existing department.
func (s *DepartmentService) UpdateDepartment(ctx context.Context, id uint, in DepartmentUpdateInput) (err error) {
	// 1. Validate ID
	if id == 0 {
		return gerror.NewCodef(gcode.CodeInvalidArgument, g.I18n().T(ctx, "department.idInvalid"))
	}

	// 2. Validate input DTO
	// The `required-without-all` rule ensures at least one field is present for update.
	if errVal := g.Validator().Data(in).Run(ctx); errVal != nil {
		return gerror.WrapCode(gcode.CodeValidationFailed, errVal)
	}

	// 3. Fetch existing department
	dept, err := s.GetDepartmentById(ctx, id) // Use existing service method
	if err != nil {
		return err // This will already be CodeNotFound or other DB error from GetDepartmentById
	}
	if dept == nil { // Should be caught by GetDepartmentById, but as a safeguard
		return gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "department.notFoundId", id))
	}

	updateData := g.Map{}

	// 4. Name uniqueness check (if name is being changed)
	if in.Name != nil && *in.Name != dept.Name {
		currentParentId := dept.ParentId
		if in.ParentId != nil { // If parent is also changing, check against new parent
			currentParentId = *in.ParentId
		}

		query := dao.Department.Ctx(ctx).Where(dao.Department.Columns().Name, *in.Name).
			Where(dao.Department.Columns().ParentId, currentParentId).
			WhereNot(dao.Department.Columns().Id, id)
		count, err := query.Count()
		if err != nil {
			return gerror.Wrapf(err, "Failed to check department name uniqueness")
		}
		if count > 0 {
			return gerror.NewCodef(gcode.CodeBusinessValidationFailed, g.I18n().Tf(ctx, "department.nameTakenInParent", *in.Name))
		}
		updateData[dao.Department.Columns().Name] = *in.Name
	}

	// 5. ParentId change checks
	if in.ParentId != nil && *in.ParentId != dept.ParentId {
		newParentId := *in.ParentId
		if newParentId == id {
			return gerror.NewCode(gcode.CodeBusinessValidationFailed, g.I18n().T(ctx, "department.cannotSetSelfAsParent"))
		}
		if newParentId != 0 { // If new parent is not root, check if it exists
			parentDept, err := s.GetDepartmentById(ctx, newParentId)
			if err != nil || parentDept == nil {
				return gerror.NewCodef(gcode.CodeBusinessValidationFailed, g.I18n().Tf(ctx, "department.parentNotFound", newParentId))
			}
		}

		// Basic circular dependency check: new parent cannot be a descendant of the current department.
		// This requires fetching all descendants of the current department.
		descendants, err := s.findAllDescendantIds(ctx, id)
		if err != nil {
			return gerror.Wrapf(err, "Failed to check for circular dependency")
		}
		for _, descendantId := range descendants {
			if newParentId == descendantId {
				return gerror.NewCode(gcode.CodeBusinessValidationFailed, g.I18n().T(ctx, "department.cannotMoveToDescendant"))
			}
		}
		updateData[dao.Department.Columns().ParentId] = newParentId
	}

	// 6. Other fields
	if in.Sort != nil && *in.Sort != dept.Sort {
		updateData[dao.Department.Columns().Sort] = *in.Sort
	}
	if in.Status != nil && (dept.Status == nil || *in.Status != *dept.Status) { // Handle if current status was nil
		updateData[dao.Department.Columns().Status] = *in.Status
	}
	if in.Remark != nil && *in.Remark != dept.Remark {
		updateData[dao.Department.Columns().Remark] = *in.Remark
	}

	// 7. If no actual changes, return nil (or a specific message/error if preferred)
	if len(updateData) == 0 {
		// return gerror.NewCode(gcode.CodeInvalidParameter, g.I18n().T(ctx, "department.updateNoChanges"))
		return nil // No changes to apply
	}

	// 8. Add UpdatedAt timestamp and perform update
	updateData[dao.Department.Columns().UpdatedAt] = gtime.Now()
	_, err = dao.Department.Ctx(ctx).Data(updateData).Where(dao.Department.Columns().Id, id).Update()
	if err != nil {
		return gerror.Wrapf(err, "Failed to update department with ID %d", id)
	}

	return nil
}

// findAllDescendantIds recursively finds all descendant IDs of a given department.
// Used for circular dependency checks.
func (s *DepartmentService) findAllDescendantIds(ctx context.Context, deptId uint) ([]uint, error) {
	var allDescendantIds []uint
	var directChildren []*entity.Department

	// Find direct children
	err := dao.Department.Ctx(ctx).Where(dao.Department.Columns().ParentId, deptId).Fields(dao.Department.Columns().Id).Scan(&directChildren)
	if err != nil {
		return nil, err
	}

	for _, child := range directChildren {
		allDescendantIds = append(allDescendantIds, child.Id)
		// Recursively find children of this child
		grandChildrenIds, err := s.findAllDescendantIds(ctx, child.Id)
		if err != nil {
			return nil, err // Propagate error up
		}
		allDescendantIds = append(allDescendantIds, grandChildrenIds...)
	}
	return allDescendantIds, nil
}

// DeleteDepartment deletes a department by its ID.
// It checks if the department has any child departments before deletion.
func (s *DepartmentService) DeleteDepartment(ctx context.Context, id uint) (err error) {
	// 1. Validate ID
	if id == 0 {
		return gerror.NewCodef(gcode.CodeInvalidArgument, g.I18n().T(ctx, "department.idInvalid"))
	}

	// 2. Check if the department exists
	dept, err := s.GetDepartmentById(ctx, id) // Use existing service method
	if err != nil {
		// This could be a DB error or CodeNotFound from GetDepartmentById
		return err
	}
	if dept == nil { // Should be caught by GetDepartmentById, but as a safeguard
		return gerror.NewCodef(gcode.CodeNotFound, g.I18n().Tf(ctx, "department.notFoundId", id))
	}

	// 3. Child Department Check
	count, err := dao.Department.Ctx(ctx).Where(dao.Department.Columns().ParentId, id).Count()
	if err != nil {
		return gerror.Wrapf(err, "Failed to check for child departments for ID %d", id)
	}
	if count > 0 {
		return gerror.NewCode(gcode.CodeBusinessValidationFailed, g.I18n().T(ctx, "department.hasChildrenCannotDelete"))
	}

	// 4. Proceed to delete the department
	_, err = dao.Department.Ctx(ctx).Where(dao.Department.Columns().Id, id).Delete()
	if err != nil {
		return gerror.Wrapf(err, "Failed to delete department with ID %d", id)
	}

	// TODO: Future consideration: what to do with users/roles associated with this department?
	// This might involve calls to other services or direct DAO operations on related tables,
	// depending on business rules (e.g., set user's department_id to null, reassign, or prevent deletion if users exist).
	// For now, only the department itself is deleted.

	return nil
}
```
