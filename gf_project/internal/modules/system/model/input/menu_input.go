package input

// MenuCreateInp is the input parameter for creating a new menu.
// Fields are based on devinggo's 'system_menu' table structure and common menu properties.
type MenuCreateInp struct {
	ParentId  uint64 `json:"parentId"  dc:"Parent Menu ID (0 for top-level menu)"`
	Name      string `json:"name"      v:"required|length:1,60#Menu name is required|Menu name length must be between 1 and 60 characters" dc:"Menu Name (display text/label)"`
	Code      string `json:"code"      v:"required|length:1,100#Menu code/identifier is required|Code length must be between 1 and 100 characters" dc:"Menu Identifier/Permission Code (e.g., system:user:list, unique)"`
	Icon      string `json:"icon,omitempty"      v:"max-length:50#Icon string too long" dc:"Menu Icon (e.g., ant-design-icon class)"`
	Route     string `json:"route,omitempty"     v:"max-length:200#Route path too long" dc:"Routing Path (frontend URL)"`
	Component string `json:"component,omitempty" v:"max-length:255#Component path too long" dc:"Component Path (frontend component file path)"`
	Redirect  string `json:"redirect,omitempty"  v:"max-length:255#Redirect path too long" dc:"Redirect path for menu item"`
	IsHidden  uint   `json:"isHidden"  v:"required|in:0,1#Is Hidden must be 0 (No) or 1 (Yes)" dc:"Is Hidden in menu (0:No, 1:Yes)"` // Changed to 0/1 for clarity, common boolean mapping
	Type      string `json:"type"      v:"required|in:M,C,F,L,I#Menu type is required|Type must be M, C, F, L, or I" dc:"Menu Type (M:Directory, C:Menu, F:Button/Function, L:External Link, I:Iframe)"`
	Status    uint   `json:"status"    v:"required|in:0,1#Status must be 0 (Normal) or 1 (Disabled)" dc:"Status (0:Normal, 1:Disabled)"` // Changed to 0/1 for clarity
	Sort      int    `json:"sort,omitempty"      dc:"Sort Order (ascending)"`
	Remark    string `json:"remark,omitempty"    v:"max-length:255#Remark too long" dc:"Optional remark"`
}

// MenuUpdateInp is the input parameter for updating an existing menu.
type MenuUpdateInp struct {
	Id        uint64 `json:"id"        v:"required|min:1#Menu ID is required" dc:"Menu ID"`
	ParentId  uint64 `json:"parentId"  dc:"Parent Menu ID (0 for top-level menu)"`
	Name      string `json:"name"      v:"required|length:1,60#Menu name is required" dc:"Menu Name"`
	Code      string `json:"code"      v:"required|length:1,100#Menu code is required" dc:"Menu Identifier/Permission Code"`
	Icon      string `json:"icon,omitempty"      v:"max-length:50" dc:"Menu Icon"`
	Route     string `json:"route,omitempty"     v:"max-length:200" dc:"Routing Path"`
	Component string `json:"component,omitempty" v:"max-length:255" dc:"Component Path"`
	Redirect  string `json:"redirect,omitempty"  v:"max-length:255" dc:"Redirect Path"`
	IsHidden  uint   `json:"isHidden"  v:"required|in:0,1" dc:"Is Hidden (0:No, 1:Yes)"`
	Type      string `json:"type"      v:"required|in:M,C,F,L,I" dc:"Menu Type"`
	Status    uint   `json:"status"    v:"required|in:0,1" dc:"Status (0:Normal, 1:Disabled)"`
	Sort      int    `json:"sort,omitempty"      dc:"Sort Order"`
	Remark    string `json:"remark,omitempty"    v:"max-length:255" dc:"Remark"`
}

// MenuListInp is for fetching menu list. Typically returns a tree structure.
// Filters might be applied before tree construction.
type MenuListInp struct {
	Name   string `json:"name,omitempty"   dc:"Filter by menu name (fuzzy match)"`
	Status *uint  `json:"status,omitempty" v:"in:0,1#Status must be 0 or 1 if provided" dc:"Filter by status (0:Normal, 1:Disabled)"`
	// No pagination for menu trees usually; all are fetched and tree is built.
}

// MenuDeleteInp for deleting a menu. Note: Deleting a menu might affect child menus.
type MenuDeleteInp struct {
	Id uint64 `json:"id" v:"required|min:1#Menu ID is required for deletion" dc:"Menu ID"`
}
