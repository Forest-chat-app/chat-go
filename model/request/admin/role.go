package admin

// ListRolesRequest 角色列表（角色表是固定数据，只做查询和修改描述）
type ListRolesRequest struct {
	Page     int `form:"page"`
	PageSize int `form:"page_size"`
}

// UpdateRoleRequest 修改角色描述
type UpdateRoleRequest struct {
	RoleId      string `json:"role_id" binding:"required"`
	Description string `json:"description" binding:"required"`
}

// AssignUserRoleRequest 分配用户系统角色（超级管理员操作普通管理员）
type AssignUserRoleRequest struct {
	UserId string `json:"user_id" binding:"required"`
	RoleId string `json:"role_id" binding:"required"`
}
