package service

import (
	"context"
	"yuncms/internal/logic/post"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"yuncms/internal/model/output"
)

type IPostService interface {
	Create(ctx context.Context, in *input.PostCreateInput) (*entity.Post, error)
	GetById(ctx context.Context, id uint64) (*entity.Post, error)
    Update(ctx context.Context, id uint64, in *input.PostUpdateInput) error
    Delete(ctx context.Context, id uint64) error
    List(ctx context.Context, in *input.PostListInput) (*output.PostListOutput, error)
}

type postServiceImpl struct {
	postLogic *post.PostLogic
}

func NewPostService() IPostService {
	return &postServiceImpl{postLogic: post.NewPostLogic()}
}

func (s *postServiceImpl) Create(ctx context.Context, in *input.PostCreateInput) (*entity.Post, error) {
    return s.postLogic.Create(ctx, in)
}
func (s *postServiceImpl) GetById(ctx context.Context, id uint64) (*entity.Post, error) {
    return s.postLogic.GetById(ctx, id)
}
func (s *postServiceImpl) Update(ctx context.Context, id uint64, in *input.PostUpdateInput) error {
    return s.postLogic.Update(ctx, id, in)
}
func (s *postServiceImpl) Delete(ctx context.Context, id uint64) error {
    return s.postLogic.Delete(ctx, id)
}
func (s *postServiceImpl) List(ctx context.Context, in *input.PostListInput) (*output.PostListOutput, error) {
    return s.postLogic.List(ctx, in)
}
