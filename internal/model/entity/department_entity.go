package entity

import "github.com/gogf/gf/v2/os/gtime"

type Department struct {
	Id        uint64      `json:"id"        orm:"id,primary"`
	ParentId  uint64      `json:"parentId"  orm:"parent_id"`
	Name      string      `json:"name"      orm:"name"`
	Leader    string      `json:"leader"    orm:"leader"`
	Status    int         `json:"status"    orm:"status"`
	SortOrder int         `json:"sortOrder" orm:"sort_order"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at"`
}
