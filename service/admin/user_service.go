package admin

import (
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	reqAdmin "chat-server/model/request/admin"
	"chat-server/utils"
)

type AdminUserService struct{}

// ListUsers 用户列表（分页+关键词搜索）
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

	for i := range users {
		users[i].Avatar = utils.GenerateCdnUrl(users[i].Avatar)
	}

	return map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"users":     users,
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

// DeleteUser 删除用户（软删除：暂时直接物理删除user_role + user，注意实际场景根据业务决定）
func (s *AdminUserService) DeleteUser(req reqAdmin.DeleteUserRequest) error {
	tx := global.CHAT_MYSQL.Begin()
	if tx.Error != nil {
		return common.NewServiceError(common.ERROR)
	}
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		} else if tx.Error != nil {
			tx.Rollback()
		} else {
			tx.Commit()
		}
	}()

	// 删除用户角色
	if err := tx.Where("user_id = ?", req.UserId).Delete(&mysql.UserRole{}).Error; err != nil {
		tx.Error = err
		return common.NewServiceError(common.ERROR)
	}
	// 删除房间成员记录
	if err := tx.Where("user_id = ?", req.UserId).Delete(&mysql.RoomMembers{}).Error; err != nil {
		tx.Error = err
		return common.NewServiceError(common.ERROR)
	}
	// 删除用户
	if err := tx.Where("id = ?", req.UserId).Delete(&mysql.User{}).Error; err != nil {
		tx.Error = err
		global.CHAT_LOG.Error("AdminDeleteUser-->删除失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}
