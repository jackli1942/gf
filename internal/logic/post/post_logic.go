package post

import (
	"context"
	"yuncms/internal/dao"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"yuncms/internal/model/output"
    "yuncms/internal/utility/i18nutil"
	"github.com/gogf/gf/v2/errors/gerror"
	// "database/sql" // Not directly used, DAO handles sql.ErrNoRows
)

type PostLogic struct {
	postDao *dao.PostDao
}

func NewPostLogic() *PostLogic {
	return &PostLogic{postDao: dao.NewPostDao()}
}

func (pl *PostLogic) Create(ctx context.Context, in *input.PostCreateInput) (*entity.Post, error) {
    existingByCode, err := pl.postDao.GetByCode(ctx, in.Code)
    if err != nil {
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db_check_post_code", "Code", in.Code)) // New i18n key
    }
    if existingByCode != nil {
        return nil, gerror.New(i18nutil.T(ctx, "post_code_exists", "Code", in.Code))
    }

	post := &entity.Post{
		Code:      in.Code,
		Name:      in.Name,
		Status:    in.Status,
		SortOrder: in.SortOrder,
		Remark:    in.Remark,
	}
	id, err := pl.postDao.Create(ctx, post)
	if err != nil {
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db_create_post"))
    }
	post.Id = uint64(id)
	return post, nil
}

func (pl *PostLogic) GetById(ctx context.Context, id uint64) (*entity.Post, error) {
	post, err := pl.postDao.GetById(ctx, id)
	if err != nil { // DAO error other than not found
        return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db"))
    }
	if post == nil { // DAO returned nil, nil for not found
        return nil, gerror.New(i18nutil.T(ctx, "post_not_found"))
    }
	return post, nil
}

func (pl *PostLogic) Update(ctx context.Context, id uint64, in *input.PostUpdateInput) error {
    post, err := pl.postDao.GetById(ctx, id)
    if err != nil { return gerror.Wrap(err, i18nutil.T(ctx, "error_db")) }
    if post == nil { return gerror.New(i18nutil.T(ctx, "post_not_found")) }

    if in.Code != nil && *in.Code != post.Code {
        existing, err := pl.postDao.GetByCode(ctx, *in.Code)
        if err != nil { return gerror.Wrap(err, i18nutil.T(ctx, "error_db"))}
        if existing != nil && existing.Id != id { // Code taken by another post
            return gerror.New(i18nutil.T(ctx, "post_code_exists", "Code", *in.Code))
        }
        post.Code = *in.Code
    }
    if in.Name != nil { post.Name = *in.Name }
    if in.Status != nil { post.Status = *in.Status }
    if in.SortOrder != nil { post.SortOrder = *in.SortOrder }
    if in.Remark != nil { post.Remark = *in.Remark }

    err = pl.postDao.Update(ctx, post)
    if err != nil {
        return gerror.Wrap(err, i18nutil.T(ctx, "error_db_update_post")) // New i18n key
    }
    return nil
}

func (pl *PostLogic) Delete(ctx context.Context, id uint64) error {
    post, err := pl.postDao.GetById(ctx, id)
    if err != nil { return gerror.Wrap(err, i18nutil.T(ctx, "error_db")) }
    if post == nil { return gerror.New(i18nutil.T(ctx, "post_not_found")) }

    // Optional: Check if post is assigned to any users before deleting.

    err = pl.postDao.Delete(ctx, id)
    if err != nil {
        return gerror.Wrap(err, i18nutil.T(ctx, "error_db_delete_post")) // New i18n key
    }
    return nil
}

func (pl *PostLogic) List(ctx context.Context, in *input.PostListInput) (*output.PostListOutput, error) {
    posts, total, err := pl.postDao.List(ctx, in)
    if err != nil { return nil, gerror.Wrap(err, i18nutil.T(ctx, "error_db")) }
    return &output.PostListOutput{ List: posts, Total: total, Page: in.Page, Size: in.PageSize, }, nil
}
