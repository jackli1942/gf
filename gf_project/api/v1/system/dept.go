package system

import (
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"github.com/gogf/gf/v2/frame/g"
)

// DeptTreeItem represents a single item in a department tree.
// It embeds entity.SystemDept and adds a 'Children' field for hierarchical structure.
// It also includes 'value' (ID) and 'label' (Name) for common tree selection components.
type DeptTreeItem struct {
	*entity.SystemDept
	Value    uint64          `json:"value" dc:"Value for tree selection (usually ID)"`
	Label    string          `json:"label" dc:"Label for tree selection (usually Name)"`
	Children []*DeptTreeItem `json:"children,omitempty" dc:"Child department items"`
}

// DeptCreateReq is the API request structure for creating a department.
type DeptCreateReq struct {
	g.Meta `path:"/dept/create" method:"post" tags:"System/DeptAdmin" summary:"Create new department" security:"BearerAuth" group:"SystemDeptAdmin"`
	input.DeptCreateInp
}

// DeptCreateRes defines the response structure for creating a department.
type DeptCreateRes struct {
	DeptId uint64 `json:"deptId" dc:"Newly created Department ID"`
}

// DeptUpdateReq for updating a department by admin.
type DeptUpdateReq struct {
	g.Meta              `path:"/dept/update" method:"put" tags:"System/DeptAdmin" summary:"Update department details (admin)" security:"BearerAuth" group:"SystemDeptAdmin"`
	input.DeptUpdateInp // Contains the ID and all updatable fields
}

// type DeptUpdateRes struct {} // Empty on success

// DeptDeleteReq for deleting a department by admin.
type DeptDeleteReq struct {
	g.Meta `path:"/dept/delete/{id}" method:"delete" tags:"System/DeptAdmin" summary:"Delete department (admin)" security:"BearerAuth" group:"SystemDeptAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#Department ID must be a positive integer" dc:"Department ID to delete"`
}

// type DeptDeleteRes struct {} // Empty on success

// DeptListReq for listing all departments (admin view, typically as a tree).
type DeptListReq struct {
	g.Meta            `path:"/dept/list" method:"get" tags:"System/DeptAdmin" summary:"List all department items (admin view, usually a tree)" security:"BearerAuth" group:"SystemDeptAdmin"`
	input.DeptListInp // Contains filters like Name, Status
}

// DeptListRes defines the response structure for listing all departments.
type DeptListRes struct {
	List []*DeptTreeItem `json:"list" dc:"List of department items, potentially hierarchical"`
}

// DeptGetReq for getting a single department's details (admin).
type DeptGetReq struct {
	g.Meta `path:"/dept/{id}" method:"get" tags:"System/DeptAdmin" summary:"Get specific department details by ID (admin)" security:"BearerAuth" group:"SystemDeptAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#Department ID must be a positive integer" dc:"Department ID"`
}

// DeptGetRes defines the response structure for getting a single department item.
type DeptGetRes struct {
	*entity.SystemDept
}
