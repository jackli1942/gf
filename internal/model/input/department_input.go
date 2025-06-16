package input

type DepartmentCreateInput struct {
	ParentId  uint64 `json:"parentId"  v:"min:0#Parent ID must be non-negative" d:"0"`
	Name      string `json:"name"      v:"required|length:2,100#Department name is required|Name length must be 2-100 chars"`
	Leader    string `json:"leader"    v:"length:0,50#Leader name too long"`
	Status    int    `json:"status"    v:"in:1,2#Invalid status" d:"1"`
	SortOrder int    `json:"sortOrder" v:"min:0#Sort order must be non-negative" d:"0"`
}

type DepartmentUpdateInput struct {
	ParentId  *uint64 `json:"parentId"  v:"min:0#Parent ID must be non-negative"`
	Name      *string `json:"name"      v:"required|length:2,100#Department name is required|Name length must be 2-100 chars"`
	Leader    *string `json:"leader"    v:"length:0,50#Leader name too long"`
	Status    *int    `json:"status"    v:"in:1,2#Invalid status"`
	SortOrder *int    `json:"sortOrder" v:"min:0#Sort order must be non-negative"`
}

type DepartmentListInput struct {
	Page     int    `json:"page" d:"1"`
	PageSize int    `json:"pageSize" d:"10"`
	Name     string `json:"name" v:"length:0,100"` // Optional filter by name
	Status   int    `json:"status" v:"in:0,1,2" d:"0"` // Optional filter by status (0 means all)
}
