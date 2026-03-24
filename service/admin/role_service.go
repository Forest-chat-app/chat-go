package admin

import (
	"chat-server/constant"
	"chat-server/global"
	"chat-server/model/common"
	"chat-server/model/mysql"
	reqAdmin "chat-server/model/request/admin"
	"chat-server/utils"
)

type AdminRoleService struct{}

// ListRoles 获取所有角色列表
func (s *AdminRoleService) ListRoles(req reqAdmin.ListRolesRequest) (map[string]interface{}, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	var total int64
	global.CHAT_MYSQL.Model(&mysql.Role{}).Count(&total)

	var roles []mysql.Role
	global.CHAT_MYSQL.Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&roles)

	return map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"roles":     roles,
	}, nil
}

// UpdateRole 修改角色（名称、描述均可改）
func (s *AdminRoleService) UpdateRole(req reqAdmin.UpdateRoleRequest) error {
	var role mysql.Role
	if err := global.CHAT_MYSQL.First(&role, "id = ?", req.RoleId).Error; err != nil {
		return common.NewServiceError(common.ROLE_NOT_FOUND)
	}

	updates := map[string]interface{}{
		"updated_at": utils.GetUTCMillisTimestamp(),
	}
	if req.Name != "" {
		updates["name"] = req.Name
	}
	if req.Description != "" {
		updates["description"] = req.Description
	}

	if err := global.CHAT_MYSQL.Model(&mysql.Role{}).Where("id = ?", req.RoleId).Updates(updates).Error; err != nil {
		global.CHAT_LOG.Error("AdminUpdateRole-->更新失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}

// CreateRole 创建角色
func (s *AdminRoleService) CreateRole(req reqAdmin.CreateRoleRequest) error {
	// 检查 role_id 是否已存在
	var count int64
	global.CHAT_MYSQL.Model(&mysql.Role{}).Where("role_id = ?", req.RoleId).Count(&count)
	if count > 0 {
		return common.NewServiceError(common.ResponseCode{Code: 421, Msg: "角色编号已存在"})
	}

	if err := global.CHAT_MYSQL.Create(&mysql.Role{
		ID:          utils.GenerateUUid(),
		Name:        req.Name,
		Description: req.Description,
		RoleId:      req.RoleId,
		CreatedAt:   int(utils.GetUTCMillisTimestamp()),
		UpdatedAt:   int(utils.GetUTCMillisTimestamp()),
	}).Error; err != nil {
		global.CHAT_LOG.Error("AdminCreateRole-->创建失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}

// DeleteRole 删除角色
func (s *AdminRoleService) DeleteRole(req reqAdmin.DeleteRoleRequest) error {
	var role mysql.Role
	if err := global.CHAT_MYSQL.First(&role, "id = ?", req.RoleId).Error; err != nil {
		return common.NewServiceError(common.ROLE_NOT_FOUND)
	}

	if err := global.CHAT_MYSQL.Where("id = ?", req.RoleId).Delete(&mysql.Role{}).Error; err != nil {
		global.CHAT_LOG.Error("AdminDeleteRole-->删除失败", "err", err)
		return common.NewServiceError(common.ERROR)
	}
	return nil
}

// ListAdmins 列出管理员（超级+普通）
func (s *AdminRoleService) ListAdmins(req reqAdmin.AdminListAdminsRequest) (map[string]interface{}, error) {
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 || req.PageSize > 100 {
		req.PageSize = 20
	}

	// 找到管理员角色ID
	var adminRoles []mysql.Role
	global.CHAT_MYSQL.Where("role_id IN ?", []int16{constant.RoleSuperAdmin, constant.RoleCommonAdmin}).Find(&adminRoles)
	roleIDs := make([]string, 0, len(adminRoles))
	for _, r := range adminRoles {
		roleIDs = append(roleIDs, r.ID)
	}

	if len(roleIDs) == 0 {
		return map[string]interface{}{"total": 0, "admins": []interface{}{}}, nil
	}

	// 查询拥有管理员角色的 user_role 记录
	var userRoles []mysql.UserRole
	global.CHAT_MYSQL.Where("role_id IN ?", roleIDs).Find(&userRoles)

	userIDs := make([]string, 0, len(userRoles))
	roleMap := make(map[string]string) // userID -> roleID
	for _, ur := range userRoles {
		userIDs = append(userIDs, ur.UserID)
		roleMap[ur.UserID] = ur.RoleID
	}

	var total int64 = int64(len(userIDs))
	var users []mysql.User
	if len(userIDs) > 0 {
		global.CHAT_MYSQL.Where("id IN ?", userIDs).
			Offset((req.Page - 1) * req.PageSize).Limit(req.PageSize).Find(&users)
	}

	// 构建角色 ID->角色 map
	roleInfoMap := make(map[string]mysql.Role)
	for _, r := range adminRoles {
		roleInfoMap[r.ID] = r
	}

	type AdminInfo struct {
		mysql.User
		Role mysql.Role `json:"role"`
	}
	admins := make([]AdminInfo, 0, len(users))
	for _, u := range users {
		u.Avatar = utils.GenerateCdnUrl(u.Avatar)
		rID := roleMap[u.ID]
		admins = append(admins, AdminInfo{User: u, Role: roleInfoMap[rID]})
	}

	return map[string]interface{}{
		"total":     total,
		"page":      req.Page,
		"page_size": req.PageSize,
		"admins":    admins,
	}, nil
}

// CreateAdmin 将普通用户提升为管理员（超级管理员操作）
func (s *AdminRoleService) CreateAdmin(req reqAdmin.AdminCreateAdminRequest) error {
	// 检查用户存在
	var user mysql.User
	if err := global.CHAT_MYSQL.First(&user, "id = ?", req.UserId).Error; err != nil {
		return common.NewServiceError(common.USER_ID_NOT_FOUND)
	}

	// 找到普通管理员角色
	var commonAdminRole mysql.Role
	global.CHAT_MYSQL.First(&commonAdminRole, "role_id = ?", constant.RoleCommonAdmin)
	if commonAdminRole.ID == "" {
		return common.NewServiceError(common.ROLE_NOT_FOUND)
	}

	// 检查是否已有角色记录
	var userRole mysql.UserRole
	global.CHAT_MYSQL.First(&userRole, "user_id = ?", req.UserId)
	if userRole.ID == "" {
		// 新建
		if err := global.CHAT_MYSQL.Create(&mysql.UserRole{
			ID:        utils.GenerateUUid(),
			UserID:    req.UserId,
			RoleID:    commonAdminRole.ID,
			CreatedAt: utils.GetUTCMillisTimestamp(),
			UpdatedAt: utils.GetUTCMillisTimestamp(),
		}).Error; err != nil {
			return common.NewServiceError(common.ERROR)
		}
	} else {
		// 更新角色
		if err := global.CHAT_MYSQL.Model(&mysql.UserRole{}).Where("user_id = ?", req.UserId).Updates(map[string]interface{}{
			"role_id":    commonAdminRole.ID,
			"updated_at": utils.GetUTCMillisTimestamp(),
		}).Error; err != nil {
			return common.NewServiceError(common.ERROR)
		}
	}
	return nil
}

// DeleteAdmin 撤销管理员权限，降级为普通用户（超级管理员操作）
func (s *AdminRoleService) DeleteAdmin(req reqAdmin.AdminDeleteAdminRequest, operatorId string) error {
	if req.UserId == operatorId {
		return common.NewServiceError(common.ADMIN_CANNOT_DELETE_SELF)
	}

	// 找到普通用户角色
	var normalRole mysql.Role
	global.CHAT_MYSQL.First(&normalRole, "role_id = ?", constant.RoleNormalUser)
	if normalRole.ID == "" {
		return common.NewServiceError(common.ROLE_NOT_FOUND)
	}

	// 验证目标是普通管理员
	var userRole mysql.UserRole
	global.CHAT_MYSQL.First(&userRole, "user_id = ?", req.UserId)
	if userRole.ID == "" {
		return common.NewServiceError(common.ADMIN_NOT_FOUND)
	}
	var targetRole mysql.Role
	global.CHAT_MYSQL.First(&targetRole, "id = ?", userRole.RoleID)
	if targetRole.RoleId != constant.RoleCommonAdmin {
		return common.NewServiceError(common.ADMIN_NOT_FOUND)
	}

	// 降级为普通用户
	if err := global.CHAT_MYSQL.Model(&mysql.UserRole{}).Where("user_id = ?", req.UserId).Updates(map[string]interface{}{
		"role_id":    normalRole.ID,
		"updated_at": utils.GetUTCMillisTimestamp(),
	}).Error; err != nil {
		return common.NewServiceError(common.ERROR)
	}
	return nil
}
