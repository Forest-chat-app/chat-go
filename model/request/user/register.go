package user

// 用户注册请求结构
type RegisterRequest struct {
	UserAccount string `form:"user_account" json:"user_account" binding:"required"`
	Password    string `form:"password" json:"password" binding:"required"`
	Platform    string `form:"platform" json:"platform" binding:"required"`
}
