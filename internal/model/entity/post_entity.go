package entity

import "github.com/gogf/gf/v2/os/gtime"

type Post struct {
	Id        uint64      `json:"id"        orm:"id,primary"`
	Code      string      `json:"code"      orm:"code"`
	Name      string      `json:"name"      orm:"name"`
	Status    int         `json:"status"    orm:"status"`
	SortOrder int         `json:"sortOrder" orm:"sort_order"`
	Remark    string      `json:"remark"    orm:"remark"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at"`
}
