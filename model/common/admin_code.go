package common

// 管理端响应码
var (
	// 管理端权限相关：700-730
	ADMIN_UNAUTHORIZED       = ResponseCode{Code: 700, Msg: "管理员未授权"}
	ADMIN_NOT_FOUND          = ResponseCode{Code: 701, Msg: "管理员不存在"}
	ADMIN_ACCOUNT_NOT_FOUND  = ResponseCode{Code: 702, Msg: "管理员账号不存在"}
	ADMIN_PASSWORD_INVALID   = ResponseCode{Code: 703, Msg: "密码错误"}
	ADMIN_FORBIDDEN          = ResponseCode{Code: 704, Msg: "权限不足"}
	ADMIN_ACCOUNT_EXISTS     = ResponseCode{Code: 705, Msg: "该账号已存在"}
	ADMIN_ONLY_SUPER         = ResponseCode{Code: 706, Msg: "仅超级管理员可执行此操作"}
	ADMIN_CANNOT_DELETE_SELF = ResponseCode{Code: 707, Msg: "不能删除自己"}

	// 消息相关：750-780
	MESSAGE_NOT_FOUND = ResponseCode{Code: 750, Msg: "消息不存在"}
)
