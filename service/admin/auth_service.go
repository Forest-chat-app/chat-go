package admin

import (
	"chat-server/constant"
	"chat-server/global"

	"chat-server/model/common"
	"chat-server/model/mysql"
	reqAdmin "chat-server/model/request/admin"
	"chat-server/utils"
)

type AdminAuthService struct{}

// Login 管理员登录
func (s *AdminAuthService) Login(req reqAdmin.AdminLoginRequest) (map[string]interface{}, error) {
	// 1、查找用户
	var queryUser mysql.User
	if err := global.CHAT_MYSQL.Where("user_account = ?", req.UserAccount).First(&queryUser).Error; err != nil {
		return nil, common.NewServiceError(common.ADMIN_ACCOUNT_NOT_FOUND)
	}

	// 2、验证密码
	match, err := utils.CompareHashAndPassword(queryUser.Password, req.Password)
	if err != nil {
		return nil, common.NewServiceError(common.ERROR)
	}
	if !match {
		return nil, common.NewServiceError(common.ADMIN_PASSWORD_INVALID)
	}

	// 3、获取用户角色，验证是否为管理员
	var userRole mysql.UserRole
	global.CHAT_MYSQL.First(&userRole, "user_id = ?", queryUser.ID)
	if userRole.ID == "" {
		return nil, common.NewServiceError(common.ADMIN_FORBIDDEN)
	}

	var role mysql.Role
	global.CHAT_MYSQL.First(&role, "id = ?", userRole.RoleID)
	if role.ID == "" {
		return nil, common.NewServiceError(common.ADMIN_FORBIDDEN)
	}

	if role.RoleId != constant.RoleSuperAdmin && role.RoleId != constant.RoleCommonAdmin {
		return nil, common.NewServiceError(common.ADMIN_FORBIDDEN)
	}

	// 4、生成管理端 token
	token, err := utils.GenerateAdminToken(queryUser.ID, queryUser.UserAccount, role.RoleId)
	if err != nil {
		global.CHAT_LOG.Error("AdminLogin-->生成token失败", "err", err)
		return nil, common.NewServiceError(common.GENERATE_TOKEN_ERROR)
	}

	// 5、返回
	queryUser.Avatar = utils.GenerateCdnUrl(queryUser.Avatar)
	return map[string]interface{}{
		"token": token,
		"user":  queryUser,
		"role":  role,
	}, nil
}

// GetSelf 获取当前管理员信息
func (s *AdminAuthService) GetSelf(claims *utils.AdminTokenClaims) (map[string]interface{}, error) {
	var queryUser mysql.User
	if err := global.CHAT_MYSQL.First(&queryUser, "id = ?", claims.UserID).Error; err != nil {
		return nil, common.NewServiceError(common.ADMIN_NOT_FOUND)
	}

	var userRole mysql.UserRole
	global.CHAT_MYSQL.First(&userRole, "user_id = ?", claims.UserID)
	var role mysql.Role
	if userRole.ID != "" {
		global.CHAT_MYSQL.First(&role, "id = ?", userRole.RoleID)
	}

	queryUser.Avatar = utils.GenerateCdnUrl(queryUser.Avatar)
	return map[string]interface{}{
		"user": queryUser,
		"role": role,
	}, nil
}
