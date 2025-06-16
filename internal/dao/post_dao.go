package dao

import (
	"context"
	"yuncms/internal/model/entity"
	"yuncms/internal/model/input"
	"github.com/gogf/gf/v2/database/gdb"
	"github.com/gogf/gf/v2/frame/g"
	"github.com/gogf/gf/v2/errors/gerror"
    "database/sql"
	"fmt" // For error formatting in Create
)

var _ gdb.DB = g.DB()

type PostDao struct{}

func NewPostDao() *PostDao { return &PostDao{} }

func (d *PostDao) Create(ctx context.Context, post *entity.Post) (id int64, err error) {
	res, err := g.DB().Model("posts").Data(post).Insert()
	if err != nil { return 0, err }
	newId, err := res.LastInsertId()
    if err != nil {
        return 0, fmt.Errorf("failed to get last insert ID for post: %w", err)
    }
	return newId, nil
}

func (d *PostDao) GetById(ctx context.Context, id uint64) (*entity.Post, error) {
	var post *entity.Post
	err := g.DB().Model("posts").Where("id", id).Scan(&post)
    if err != nil {
        if gerror.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
	return post, nil
}

func (d *PostDao) GetByCode(ctx context.Context, code string) (*entity.Post, error) {
	var post *entity.Post
	err := g.DB().Model("posts").Where("code", code).Scan(&post)
    if err != nil {
        if gerror.Is(err, sql.ErrNoRows) {
            return nil, nil
        }
        return nil, err
    }
	return post, nil
}

func (d *PostDao) Update(ctx context.Context, post *entity.Post) error {
    updateData := g.Map{
        "code":       post.Code,
        "name":       post.Name,
        "status":     post.Status,
        "sort_order": post.SortOrder,
        "remark":     post.Remark,
    }
	_, err := g.DB().Model("posts").Data(updateData).Where("id", post.Id).Update()
	return err
}

func (d *PostDao) Delete(ctx context.Context, id uint64) error {
	_, err := g.DB().Model("posts").Where("id", id).Delete()
	return err
}

func (d *PostDao) List(ctx context.Context, in *input.PostListInput) (posts []*entity.Post, total int, err error) {
    query := g.DB().Model("posts")
    if in.Name != "" { query = query.Where("name LIKE ?", "%"+in.Name+"%") }
    if in.Code != "" { query = query.Where("code LIKE ?", "%"+in.Code+"%") }
    if in.Status != 0 { query = query.Where("status", in.Status) }

    total, err = query.Count()
    if err != nil { return nil, 0, gerror.Wrap(err, "PostDao: Count failed") }
    if total == 0 { return []*entity.Post{}, 0, nil }

    err = query.Page(in.Page, in.PageSize).Order("sort_order ASC, id ASC").Scan(&posts)
    if err != nil { return nil, 0, gerror.Wrap(err, "PostDao: Scan failed") }
    return posts, total, nil
}
