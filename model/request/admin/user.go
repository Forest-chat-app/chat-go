package admin

// ListUsersRequest 用户列表请求
type ListUsersRequest struct {
	Page     int    `form:"page"`
	PageSize int    `form:"page_size"`
	Keyword  string `form:"keyword"`
}

// UpdateAdminUserProfileRequest 管理端修改用户基本信息（不含头像）
type UpdateAdminUserProfileRequest struct {
	UserId      string `json:"user_id" binding:"required"`
	UserAccount string `json:"user_account"`
	Nickname    string `json:"nickname"`
	Email       string `json:"email"`
}

// UpdateAdminUserAvatarRequest 管理端修改用户头像
type UpdateAdminUserAvatarRequest struct {
	UserId string `json:"user_id" binding:"required"`
	Avatar string `json:"avatar" binding:"required"`
}

// ResetUserPasswordRequest 重置用户密码
type ResetUserPasswordRequest struct {
	UserId      string `json:"user_id" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// BanUserRequest 封禁用户
type BanUserRequest struct {
	UserId string `json:"user_id" binding:"required"`
}

// GetUserDetailRequest 获取用户详情
type GetUserDetailRequest struct {
	UserId string `form:"user_id" binding:"required"`
}
