package service

import (
	"context"
	"yuncms/internal/logic/user"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
)

type IUserService interface {
	Create(ctx context.Context, in *input.UserCreateInput) (*entity.User, error)
    GetByUsername(ctx context.Context, username string) (*entity.User, error)
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
