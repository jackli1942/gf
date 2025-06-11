// =================================================================================
// Code generated and maintained by GoFrame CLI tool. DO NOT EDIT.
// =================================================================================

package do

import (
	"github.com/gogf/gf/v2/frame/g"
)

// Menu is the golang structure of table system_menu for DAO operations like Where/Data.
type Menu struct {
	g.Meta    `orm:"table:system_menu, do:true"`
	Id        interface{} //
	ParentId  interface{} //
	Level     interface{} //
	Name      interface{} //
	Code      interface{} //
	Icon      interface{} //
	Route     interface{} //
	Component interface{} //
	Redirect  interface{} //
	IsHidden  interface{} //
	Type      interface{} //
	Status    interface{} //
	Sort      interface{} //
	CreatedBy interface{} //
	UpdatedBy interface{} //
	CreatedAt interface{} //
	UpdatedAt interface{} //
	DeletedAt interface{} //
	Remark    interface{} //
}
