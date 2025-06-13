package entity

import (
	"github.com/gogf/gf/v2/os/gtime"
)

// Position is the entity representing a position/role.
type Position struct {
	Id        uint        `json:"id"        orm:"id,primary" description:"Position ID"`
	Name      string      `json:"name"      orm:"name"       v:"required|length:1,100#position.nameRequired|position.nameLength" description:"Position Name"`
	Code      string      `json:"code,omitempty" orm:"code"  v:"length:1,100#position.codeLength" description:"Position Code (optional, should be unique if provided)"`
	Sort      int         `json:"sort"      orm:"sort"       description:"Sort order; smaller is higher priority" default:"0"`
	Status    int         `json:"status"    orm:"status"     v:"in:0,1#position.statusInvalid"    description:"Status (1:active, 0:disabled)" default:"1"`
	Remark    string      `json:"remark,omitempty" orm:"remark" v:"max-length:255#position.remarkMaxLength" description:"Optional remarks"`
	CreatedAt *gtime.Time `json:"createdAt" orm:"created_at" description:"Creation time"`
	UpdatedAt *gtime.Time `json:"updatedAt" orm:"updated_at" description:"Last update time"`
}
```
