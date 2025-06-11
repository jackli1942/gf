// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Role is the golang structure for table role.
type Role struct {
	Id        int       `json:"id"        orm:"id"         description:""` //
	Name      string    `json:"name"      orm:"name"       description:""` //
	Code      string    `json:"code"      orm:"code"       description:""` //
	DataScope int       `json:"dataScope" orm:"data_scope" description:""` //
	Status    int       `json:"status"    orm:"status"     description:""` //
	Sort      int       `json:"sort"      orm:"sort"       description:""` //
	CreatedBy int       `json:"createdBy" orm:"created_by" description:""` //
	UpdatedBy int       `json:"updatedBy" orm:"updated_by" description:""` //
	CreatedAt time.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt time.Time `json:"updatedAt" orm:"updated_at" description:""` //
	DeletedAt time.Time `json:"deletedAt" orm:"deleted_at" description:""` //
	Remark    string    `json:"remark"    orm:"remark"     description:""` //
}
