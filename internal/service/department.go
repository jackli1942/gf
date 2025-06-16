package service

import (
	"context"
	"yuncms/internal/logic/department"
	"yuncms/internal/model/input"
	"yuncms/internal/model/output"
    "yuncms/internal/model/entity"
)

type IDepartmentService interface {
	Create(ctx context.Context, in *input.DepartmentCreateInput) (*entity.Department, error)
	GetById(ctx context.Context, id uint64) (*output.DepartmentOutput, error)
    Update(ctx context.Context, id uint64, in *input.DepartmentUpdateInput) error
    Delete(ctx context.Context, id uint64) error
    List(ctx context.Context, in *input.DepartmentListInput) (*output.DepartmentListOutput, error)
    GetTree(ctx context.Context) ([]*output.DepartmentOutput, error)
}

type departmentServiceImpl struct {
	deptLogic *department.DepartmentLogic
}

func NewDepartmentService() IDepartmentService {
	return &departmentServiceImpl{deptLogic: department.NewDepartmentLogic()}
}

func (s *departmentServiceImpl) Create(ctx context.Context, in *input.DepartmentCreateInput) (*entity.Department, error) {
	return s.deptLogic.Create(ctx, in)
}
func (s *departmentServiceImpl) GetById(ctx context.Context, id uint64) (*output.DepartmentOutput, error) {
	return s.deptLogic.GetById(ctx, id)
}
func (s *departmentServiceImpl) Update(ctx context.Context, id uint64, in *input.DepartmentUpdateInput) error {
    return s.deptLogic.Update(ctx, id, in)
}
func (s *departmentServiceImpl) Delete(ctx context.Context, id uint64) error {
    return s.deptLogic.Delete(ctx, id)
}
func (s *departmentServiceImpl) List(ctx context.Context, in *input.DepartmentListInput) (*output.DepartmentListOutput, error) {
    return s.deptLogic.List(ctx, in)
}
func (s *departmentServiceImpl) GetTree(ctx context.Context) ([]*output.DepartmentOutput, error) {
    return s.deptLogic.GetTree(ctx)
}
