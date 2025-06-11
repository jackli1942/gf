// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Post is the golang structure for table post.
type Post struct {
	Id        int       `json:"id"        orm:"id"         description:""` //
	Name      string    `json:"name"      orm:"name"       description:""` //
	Code      string    `json:"code"      orm:"code"       description:""` //
	Sort      int       `json:"sort"      orm:"sort"       description:""` //
	Status    int       `json:"status"    orm:"status"     description:""` //
	CreatedBy int       `json:"createdBy" orm:"created_by" description:""` //
	UpdatedBy int       `json:"updatedBy" orm:"updated_by" description:""` //
	CreatedAt time.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt time.Time `json:"updatedAt" orm:"updated_at" description:""` //
	DeletedAt time.Time `json:"deletedAt" orm:"deleted_at" description:""` //
	Remark    string    `json:"remark"    orm:"remark"     description:""` //
}
