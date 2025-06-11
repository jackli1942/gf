package system

import (
	"gf_project/internal/modules/system/model/entity"
	"gf_project/internal/modules/system/model/input"
	"github.com/gogf/gf/v2/frame/g"
)

// MenuTreeItem represents a single item in a menu tree.
// It embeds entity.SystemMenu and adds a 'Children' field for hierarchical structure.
type MenuTreeItem struct {
	*entity.SystemMenu                 // Embed SystemMenu fields (Id, ParentId, Name, Code, Type, etc.)
	Children           []*MenuTreeItem `json:"children,omitempty" dc:"Child menu items"`
}

// MenuCreateReq is the API request structure for creating a menu.
type MenuCreateReq struct {
	g.Meta `path:"/menu/create" method:"post" tags:"System/MenuAdmin" summary:"Create new menu item" security:"BearerAuth" group:"SystemMenuAdmin"`
	input.MenuCreateInp
}

// MenuCreateRes defines the response structure for creating a menu.
type MenuCreateRes struct {
	MenuId uint64 `json:"menuId" dc:"Newly created Menu ID"`
}

// MenuUpdateReq for updating a menu by admin.
type MenuUpdateReq struct {
	g.Meta              `path:"/menu/update" method:"put" tags:"System/MenuAdmin" summary:"Update menu item details (admin)" security:"BearerAuth" group:"SystemMenuAdmin"`
	input.MenuUpdateInp // Contains the ID and all updatable fields
}

// type MenuUpdateRes struct {} // Empty on success typically

// MenuDeleteReq for deleting a menu by admin.
type MenuDeleteReq struct {
	g.Meta `path:"/menu/delete/{id}" method:"delete" tags:"System/MenuAdmin" summary:"Delete menu item (admin)" security:"BearerAuth" group:"SystemMenuAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#Menu ID must be a positive integer" dc:"Menu ID to delete"`
}

// type MenuDeleteRes struct {} // Empty on success

// MenuListReq for listing all menus (admin view, typically as a tree).
type MenuListReq struct {
	g.Meta            `path:"/menu/list" method:"get" tags:"System/MenuAdmin" summary:"List all menu items (admin view, usually a tree)" security:"BearerAuth" group:"SystemMenuAdmin"`
	input.MenuListInp // Contains filters like Name, Status
}

// MenuListRes defines the response structure for listing all menus.
type MenuListRes struct {
	List []*MenuTreeItem `json:"list" dc:"List of menu items, potentially hierarchical"`
}

// UserMenuListReq for getting menus accessible by the current authenticated user.
type UserMenuListReq struct {
	g.Meta `path:"/menu/user-menus" method:"get" tags:"System/Menu" summary:"Get accessible menu tree for current logged-in user" security:"BearerAuth" group:"SystemMenuUser"`
	// No specific input parameters other than what's derived from the user's JWT (e.g., user ID, roles)
}

// UserMenuListRes defines the response structure for user-specific menus.
type UserMenuListRes struct {
	List []*MenuTreeItem `json:"list" dc:"Accessible menu tree for the user"`
}

// MenuGetReq for getting a single menu's details (admin).
type MenuGetReq struct {
	g.Meta `path:"/menu/{id}" method:"get" tags:"System/MenuAdmin" summary:"Get specific menu item details by ID (admin)" security:"BearerAuth" group:"SystemMenuAdmin"`
	Id     uint64 `in:"path" v:"required|min:1#Menu ID must be a positive integer" dc:"Menu ID"`
}

// MenuGetRes defines the response structure for getting a single menu item.
type MenuGetRes struct {
	*entity.SystemMenu // Returns the flat entity. Frontend can place it in a tree if needed or use it to update a node.
}
