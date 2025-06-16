package service

import (
	"context"
	"yuncms/internal/logic/user"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"yuncms/internal/model/output"
)

type IUserService interface {
	Create(ctx context.Context, in *input.UserCreateInput) (*entity.User, error)
    GetByUsername(ctx context.Context, username string) (*entity.User, error)
    GetById(ctx context.Context, id uint64) (*entity.User, error)
    Update(ctx context.Context, id uint64, in *input.UserUpdateInput) error
    Delete(ctx context.Context, id uint64) error
    List(ctx context.Context, in *input.UserListInput) (*output.UserListOutput, error)
}

type userServiceImpl struct {
	userLogic *user.UserLogic
}

func NewUserService() IUserService {
	return &userServiceImpl{userLogic: user.NewUserLogic()}
}

func (s *userServiceImpl) Create(ctx context.Context, in *input.UserCreateInput) (*entity.User, error) {
	return s.userLogic.CreateUser(ctx, in)
}

func (s *userServiceImpl) GetByUsername(ctx context.Context, username string) (*entity.User, error) {
    return s.userLogic.GetUserByUsername(ctx, username)
}

func (s *userServiceImpl) GetById(ctx context.Context, id uint64) (*entity.User, error) {
    return s.userLogic.GetUserById(ctx, id)
}

func (s *userServiceImpl) Update(ctx context.Context, id uint64, in *input.UserUpdateInput) error {
	return s.userLogic.UpdateUser(ctx, id, in)
}

func (s *userServiceImpl) Delete(ctx context.Context, id uint64) error {
	return s.userLogic.DeleteUser(ctx, id)
}

func (s *userServiceImpl) List(ctx context.Context, in *input.UserListInput) (*output.UserListOutput, error) {
	return s.userLogic.ListUsers(ctx, in)
}
