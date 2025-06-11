// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package entity

import (
	"time"
)

// Menu is the golang structure for table menu.
type Menu struct {
	Id        int       `json:"id"        orm:"id"         description:""` //
	ParentId  int       `json:"parentId"  orm:"parent_id"  description:""` //
	Level     string    `json:"level"     orm:"level"      description:""` //
	Name      string    `json:"name"      orm:"name"       description:""` //
	Code      string    `json:"code"      orm:"code"       description:""` //
	Icon      string    `json:"icon"      orm:"icon"       description:""` //
	Route     string    `json:"route"     orm:"route"      description:""` //
	Component string    `json:"component" orm:"component"  description:""` //
	Redirect  string    `json:"redirect"  orm:"redirect"   description:""` //
	IsHidden  int       `json:"isHidden"  orm:"is_hidden"  description:""` //
	Type      string    `json:"type"      orm:"type"       description:""` //
	Status    int       `json:"status"    orm:"status"     description:""` //
	Sort      int       `json:"sort"      orm:"sort"       description:""` //
	CreatedBy int       `json:"createdBy" orm:"created_by" description:""` //
	UpdatedBy int       `json:"updatedBy" orm:"updated_by" description:""` //
	CreatedAt time.Time `json:"createdAt" orm:"created_at" description:""` //
	UpdatedAt time.Time `json:"updatedAt" orm:"updated_at" description:""` //
	DeletedAt time.Time `json:"deletedAt" orm:"deleted_at" description:""` //
	Remark    string    `json:"remark"    orm:"remark"     description:""` //
}
