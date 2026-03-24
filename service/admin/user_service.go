package admin

import (
	"chat-server/constant"
	"chat-server/core"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	reqAdmin "chat-server/model/request/admin"
	"chat-server/utils"
	"context"
	"fmt"
)

type AdminUserService struct{}

// ListUsers 用户列表（分页+关键词搜索，携带角色信息）
func (s *AdminUserService) ListUsers(req reqAdmin.ListUsersRequest) (map[string]interface{}, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	tx := global.CHAT_MYSQL.Model(&mysql.User{})
	if req.Keyword != "" {
		like := "%" + req.Keyword + "%"
		tx = tx.Where("user_account LIKE ? OR nickname LIKE ? OR email LIKE ?", like, like, like)
	}

	var total int64
	tx.Count(&total)

	var users []mysql.User
	tx.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&users)

	// 批量查角色
	userIDs := make([]string, 0, len(users))
	for i := range users {
		users[i].Avatar = utils.GenerateCdnUrl(users[i].Avatar)
		userIDs = append(userIDs, users[i].ID)
	}

	var userRoles []mysql.UserRole
	if len(userIDs) > 0 {
		global.CHAT_MYSQL.Where("user_id IN ?", userIDs).Find(&userRoles)
	}
	// userID -> roleID
	roleIDByUser := make(map[string]string, len(userRoles))
	roleIDs := make([]string, 0, len(userRoles))
	for _, ur := range userRoles {
		roleIDByUser[ur.UserID] = ur.RoleID
		roleIDs = append(roleIDs, ur.RoleID)
	}
	var roles []mysql.Role
	if len(roleIDs) > 0 {
		global.CHAT_MYSQL.Where("id IN ?", roleIDs).Find(&roles)
	}
	roleByID := make(map[string]mysql.Role, len(roles))
	for _, r := range roles {
		roleByID[r.ID] = r
	}

	type UserWithRole struct {
		mysql.User
		Role mysql.Role `json:"role"`
	}
	result := make([]UserWithRole, 0, len(users))
	for _, u := range users {
		rID := roleIDByUser[u.ID]
		result = append(result, UserWithRole{User: u, Role: roleByID[rID]})
	}

	return map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"users":     result,
	}, nil
}

// GetUserDetail 获取用户详情（含角色、房间信息）
func (s *AdminUserService) GetUserDetail(userId string) (map[string]interface{}, error) {
	var queryUser mysql.User
	if err := global.CHAT_MYSQL.First(&queryUser, "id = ?", userId).Error; err != nil {
		return nil, common.NewServiceError(common.USER_ID_NOT_FOUND)
	}
	queryUser.Avatar = utils.GenerateCdnUrl(queryUser.Avatar)

	var userRole mysql.UserRole
	global.CHAT_MYSQL.First(&userRole, "user_id = ?", userId)
	var role mysql.Role
	if userRole.ID != "" {
		global.CHAT_MYSQL.First(&role, "id = ?", userRole.RoleID)
	}

	var roomIDs []string
	global.CHAT_MYSQL.Model(&mysql.RoomMembers{}).Where("user_id = ?", userId).Pluck("room_id", &roomIDs)
	if roomIDs == nil {
		roomIDs = []string{}
	}

	return map[string]interface{}{
		"user":     queryUser,
		"role":     role,
		"room_ids": roomIDs,
	}, nil
}

// UpdateUserProfile 修改用户基本信息（不含头像）
func (s *AdminUserService) UpdateUserProfile(req reqAdmin.UpdateAdminUserProfileRequest) error {
	updates := map[string]interface{}{
		"updated_at": utils.GetUTCMillisTimestamp(),
	}

	if req.UserAccount != "" {
		var count int64
		global.CHAT_MYSQL.Model(&mysql.User{}).Where("user_account = ? AND id != ?", req.UserAccount, req.UserId).Count(&count)
		if count > 0 {
			return common.NewServiceError(common.USER_ACCOUNT_DUPLICATE)
		}
		updates["user_account"] = req.UserAccount
	}
	if req.Nickname != "" {
		updates["nickname"] = req.Nickname
	}
	if req.Email != "" {
		var count int64
		global.CHAT_MYSQL.Model(&mysql.User{}).Where("email = ? AND id != ?", req.Email, req.UserId).Count(&count)
		if count > 0 {
			return common.NewServiceError(common.EMAIL_DUPLICATE)
		}
		updates["email"] = req.Email
	}

	if err := global.CHAT_MYSQL.Model(&mysql.User{}).Where("id = ?", req.UserId).Updates(updates).Error; err != nil {
		global.CHAT_LOG.Error("AdminUpdateUserProfile-->更新失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}

// UpdateUserAvatar 修改用户头像
func (s *AdminUserService) UpdateUserAvatar(req reqAdmin.UpdateAdminUserAvatarRequest) error {
	if err := global.CHAT_MYSQL.Model(&mysql.User{}).Where("id = ?", req.UserId).Updates(map[string]interface{}{
		"avatar":     req.Avatar,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		global.CHAT_LOG.Error("AdminUpdateUserAvatar-->更新失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}

// ResetUserPassword 重置用户密码
func (s *AdminUserService) ResetUserPassword(req reqAdmin.ResetUserPasswordRequest) error {
	hashed, err := utils.GenerateFromPassword(req.NewPassword)
	if err != nil || hashed == "" {
		return common.NewServiceError(common.ERROR)
	}
	if err := global.CHAT_MYSQL.Model(&mysql.User{}).Where("id = ?", req.UserId).Updates(map[string]interface{}{
		"password":   hashed,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		global.CHAT_LOG.Error("AdminResetUserPassword-->更新失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}

// BanUser 封禁用户（将状态改为2，并踢出所有在线连接）
func (s *AdminUserService) BanUser(req reqAdmin.BanUserRequest) error {
	// 1、将用户状态改为封禁
	if err := global.CHAT_MYSQL.Model(&mysql.User{}).Where("id = ?", req.UserId).Updates(map[string]interface{}{
		"status":     constant.UserStatusBanned,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		global.CHAT_LOG.Error("AdminBanUser-->更新状态失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}

	// 2、踢出该用户所有WebSocket连接
	manager := global.CHAT_WEBSOCKET_MANAGER.(*core.WebSocketManager)
	manager.UserLogoutByUserId(req.UserId)

	// 3、撤销该用户在Redis中的所有token（扫描 refresh_token:{userID}:* 并删除）
	ctx := context.Background()
	pattern := fmt.Sprintf("%s:%s:*", constant.RefreshTokenPrefix, req.UserId)
	keys, err := global.CHAT_REDIS.Keys(ctx, pattern).Result()
	if err != nil {
		global.CHAT_LOG.Error("AdminBanUser-->扫描token失败", "err", err)
		return nil // 状态已改，token撤销失败不影响主流程
	}
	if len(keys) > 0 {
		global.CHAT_REDIS.Del(ctx, keys...)
	}

	return nil
}
