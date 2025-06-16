package input

type PostCreateInput struct {
	Code      string `json:"code"      v:"required|length:2,50#Post code is required|Code length 2-50"`
	Name      string `json:"name"      v:"required|length:2,100#Post name is required|Name length 2-100"`
	Status    int    `json:"status"    v:"in:1,2#Invalid status" d:"1"`
	SortOrder int    `json:"sortOrder" v:"min:0#Sort order non-negative" d:"0"`
	Remark    string `json:"remark"    v:"length:0,255#Remark too long"`
}

type PostUpdateInput struct {
	Code      *string `json:"code"      v:"required|length:2,50#Post code is required|Code length 2-50"` // Usually code is not updatable, but included if needed
	Name      *string `json:"name"      v:"required|length:2,100#Post name is required|Name length 2-100"`
	Status    *int    `json:"status"    v:"in:1,2#Invalid status"`
	SortOrder *int    `json:"sortOrder" v:"min:0#Sort order non-negative"`
	Remark    *string `json:"remark"    v:"length:0,255#Remark too long"`
}

type PostListInput struct {
	Page     int    `json:"page" d:"1"`
	PageSize int    `json:"pageSize" d:"10"`
	Name     string `json:"name" v:"length:0,100"` // Optional filter
	Code     string `json:"code" v:"length:0,50"`  // Optional filter
	Status   int    `json:"status" v:"in:0,1,2" d:"0"` // 0 for all
}
