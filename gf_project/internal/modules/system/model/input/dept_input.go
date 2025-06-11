package input

// DeptCreateInp is the input parameter for creating a new department.
// Based on devinggo's system_dept table: parent_id, name, leader, phone, status, sort, remark.
// 'level' and 'ancestors' are typically calculated by the service.
type DeptCreateInp struct {
	ParentId uint64 `json:"parentId" dc:"Parent Department ID (0 for top-level department)"`
	Name     string `json:"name"     v:"required|length:1,30#Department name is required|Department name length must be between 1 and 30 characters" dc:"Department Name"`
	Leader   string `json:"leader,omitempty"   v:"max-length:20#Leader name cannot exceed 20 characters" dc:"Leader Name"`
	Phone    string `json:"phone,omitempty"    v:"phone|max-length:20#Invalid phone number format or too long" dc:"Contact Phone Number"` // Max length 20 to be safe for various formats
	Status   uint   `json:"status"   v:"required|in:0,1#Status must be 0 (Normal) or 1 (Disabled)" dc:"Status (0:Normal, 1:Disabled)"`    // Standardized to 0/1
	Sort     int    `json:"sort,omitempty"     dc:"Sort Order (ascending)"`
	Remark   string `json:"remark,omitempty"   v:"max-length:255#Remark cannot exceed 255 characters" dc:"Optional remark"`
}

// DeptUpdateInp is the input parameter for updating an existing department.
type DeptUpdateInp struct {
	Id       uint64 `json:"id"        v:"required|min:1#Department ID is required and must be a positive integer" dc:"Department ID"`
	ParentId uint64 `json:"parentId" dc:"Parent Department ID"`
	Name     string `json:"name"     v:"required|length:1,30#Department name is required" dc:"Department Name"`
	Leader   string `json:"leader,omitempty"   v:"max-length:20" dc:"Leader Name"`
	Phone    string `json:"phone,omitempty"    v:"phone|max-length:20" dc:"Contact Phone"`
	Status   uint   `json:"status"   v:"required|in:0,1" dc:"Status (0:Normal, 1:Disabled)"`
	Sort     int    `json:"sort,omitempty"     dc:"Sort Order"`
	Remark   string `json:"remark,omitempty"   v:"max-length:255" dc:"Remark"`
}

// DeptListInp is for fetching department list, typically returned as a tree or flat list for selection.
// Filters might be applied before tree construction or on the flat list.
type DeptListInp struct {
	Name   string `json:"name,omitempty"   dc:"Filter by department name (fuzzy match)"`
	Status *uint  `json:"status,omitempty" v:"in:0,1#Status must be 0 or 1 if provided" dc:"Filter by status (0:Normal, 1:Disabled)"`
	// No pagination for department trees usually; all are fetched and tree is built,
	// or if a flat list for selection, pagination might be added later if needed.
}

// DeptDeleteInp for deleting a department. Note: Deleting a department might affect child departments and users within.
type DeptDeleteInp struct {
	Id uint64 `json:"id" v:"required|min:1#Department ID is required for deletion" dc:"Department ID"`
}
