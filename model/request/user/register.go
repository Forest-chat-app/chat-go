package user

// 用户注册请求结构
type RegisterRequest struct {
	UserAccount string `form:"user_account" json:"user_account" binding:"required"`
	Password    string `form:"password" json:"password" binding:"required"`
	Email       string `form:"email" json:"email" binding:"omitempty,email"`
	Avatar      string `form:"avatar" json:"avatar" binding:"omitempty,url"`
	Platform    string `form:"platform" json:"platform" binding:"required"`
}
