package constant

const (
	RoleSuperAdmin  = 1001
	RoleCommonAdmin = 1002
	RoleNormalUser  = 1003
	RoleGroupOwner  = 1004
	RoleGroupAdmin  = 1005
	RoleGroupMember = 1006
)

var RoleNameMap = map[int16]string{
	RoleSuperAdmin:  "超级管理员",
	RoleCommonAdmin: "普通管理员",
	RoleNormalUser:  "普通用户",
	RoleGroupOwner:  "群主",
	RoleGroupAdmin:  "群管理员",
	RoleGroupMember: "群成员",
}
