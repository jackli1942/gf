package service

import (
	"context"
	"yuncms/internal/app/dao" // Corrected path
	"yuncms/internal/app/model/entity" // Corrected path
	// "yuncms/internal/app/model/do" // DO not directly used in Create method

	"github.com/gogf/gf/v2/errors/gcode"
	"github.com/gogf/gf/v2/errors/gerror"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/os/gtime"
)

// DepartmentService handles department-related business logic.
type DepartmentService struct{}

// NewDepartmentService creates and returns a new DepartmentService.
func NewDepartmentService() *DepartmentService {
	return &DepartmentService{}
}

// DepartmentCreateInput is the DTO for creating a new department.
type DepartmentCreateInput struct {
	ParentId uint   `json:"parentId" v:"min:0#department.parentIdMin" description:"Parent Department ID (0 for root)"`
	Name     string `json:"name" v:"required|length:1,100#department.nameRequired|department.nameLength" description:"Department Name"`
	Sort     int    `json:"sort" default:"0" description:"Sort order"`
	Status   *int   `json:"status" v:"in:0,1#department.statusInvalid" description:"Status (1:active, 0:disabled), defaults to 1 (active)"`
	Remark   string `json:"remark,omitempty" v:"max-length:255#department.remarkMaxLength" description:"Optional remarks"`
}

// DepartmentCreateOutput is the DTO for the result of creating a new department.
type DepartmentCreateOutput struct {
	DepartmentId uint `json:"departmentId"`
}

// CreateDepartment handles the logic for creating a new department.
func (s *DepartmentService) CreateDepartment(ctx context.Context, in DepartmentCreateInput) (*DepartmentCreateOutput, error) {
	// 1. Validate input
	if err := g.Validator().Data(in).Run(ctx); err != nil {
		return nil, err
	}

	// 2. Check for duplicate department name under the same parent
	count, err := dao.Department.Ctx(ctx).Where(dao.Department.Columns().ParentId, in.ParentId).Where(dao.Department.Columns().Name, in.Name).Count()
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to check department name existence")
	}
	if count > 0 {
		return nil, gerror.NewCodef(gcode.CodeBusinessValidationFailed, g.I18n().Tf(ctx, "department.nameTakenInParent", in.Name))
	}

	// 3. Ensure ParentId is valid (exists if not 0)
	if in.ParentId != 0 {
		var parentEntity *entity.Department
		err := dao.Department.Ctx(ctx).WherePri(in.ParentId).Scan(&parentEntity)
		if err != nil {
			return nil, gerror.Wrapf(err, "Error validating parent department ID %d", in.ParentId)
		}
		if parentEntity == nil {
			return nil, gerror.NewCodef(gcode.CodeBusinessValidationFailed, g.I18n().Tf(ctx, "department.parentNotFound", in.ParentId))
		}
	}

	// 4. Populate department entity
	now := gtime.Now()
	deptEntity := &entity.Department{
		ParentId:  in.ParentId,
		Name:      in.Name,
		Sort:      in.Sort,
		Remark:    in.Remark,
		CreatedAt: now,
		UpdatedAt: now,
	}

	if in.Status == nil {
		deptEntity.Status = 1 // Default to active
	} else {
		deptEntity.Status = *in.Status
	}

	// 5. Call DAO to create department
	result, err := dao.Department.Ctx(ctx).Data(deptEntity).Insert()
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to create department in database")
	}

	lastInsertId, err := result.LastInsertId()
	if err != nil {
		return nil, gerror.Wrap(err, "Failed to retrieve last insert ID for department")
	}

	return &DepartmentCreateOutput{DepartmentId: uint(lastInsertId)}, nil
}
