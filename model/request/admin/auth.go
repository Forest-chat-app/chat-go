package admin

// AdminLoginRequest 管理员登录请求
type AdminLoginRequest struct {
	UserAccount string `json:"user_account" binding:"required"`
	Password    string `json:"password" binding:"required"`
}
