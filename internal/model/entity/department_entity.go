package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Department is the an entity representing a department.
type Department struct {
	Id        uint        `json:"id"        orm:"id,primary" description:"Department ID"`
	ParentId  uint        `json:"parentId"  orm:"parent_id"  description:"Parent Department ID (0 for root)"`
	Name      string      `json:"name"      orm:"name"       v:"required#department.nameRequired" description:"Department Name"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Sort order; smaller is higher priority" default:"0"`
	Status    *int        `json:"status"    orm:"status"     v:"in:0,1#department.statusInvalid"  description:"Status (1:active, 0:disabled)" default:"1"`
	Remark    string      `json:"remark"    orm:"remark"     description:"Optional remarks for the department"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Last update time"`
}
```
