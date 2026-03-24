package admin

// ListRolesRequest 角色列表
type ListRolesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// CreateRoleRequest 创建角色
type CreateRoleRequest struct {
	Name        string `json:"name" binding:"required"`
	Description string `json:"description" binding:"required"`
	RoleId      int16  `json:"role_id" binding:"required"`
}

// UpdateRoleRequest 修改角色（名称、描述均可改）
type UpdateRoleRequest struct {
	RoleId      string `json:"role_id" binding:"required"` // 数据库主键 id
	Name        string `json:"name"`
	Description string `json:"description"`
}

// DeleteRoleRequest 删除角色
type DeleteRoleRequest struct {
	RoleId string `json:"role_id" binding:"required"`
}

// AssignUserRoleRequest 分配用户系统角色
type AssignUserRoleRequest struct {
	UserId string `json:"user_id" binding:"required"`
	RoleId string `json:"role_id" binding:"required"`
}
