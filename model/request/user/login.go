package user

// 用户登录请求结构
type LoginRequest struct {
	UserAccount string `json:"user_account" binding:"required"`
	Password    string `json:"password" binding:"required"`
	Platform    string `form:"platform" json:"platform" binding:"required"` //web、mobile、pad
}
