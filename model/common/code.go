package common

// ResponseCode 定义响应码和消息的组合
type ResponseCode struct {
	Code int
	Msg  string
}

// 预定义响应状态
var (
	SUCCESS        = ResponseCode{Code: 200, Msg: "操作成功"}
	ERROR          = ResponseCode{Code: 500, Msg: "操作失败"}
	INVALID_PARAMS = ResponseCode{Code: 400, Msg: "请求参数错误"}
	// 用户相关：401-430
	USER_NOT_FOUND         = ResponseCode{Code: 401, Msg: "用户不存在"}
	USER_ID_NOT_FOUND      = ResponseCode{Code: 402, Msg: "用户id不存在"}
	USER_ACCOUNT_EXISTS    = ResponseCode{Code: 403, Msg: "用户账号已存在"}
	USER_ACCOUNT_NOT_FOUND = ResponseCode{Code: 404, Msg: "用户账号不存在"}
	PASSWORD_INVALID       = ResponseCode{Code: 405, Msg: "密码错误"}
	EMAIL_INVALID          = ResponseCode{Code: 406, Msg: "邮箱错误"}
	AVATAR_INVALID         = ResponseCode{Code: 407, Msg: "头像格式错误"}
	ROLE_NOT_FOUND         = ResponseCode{Code: 408, Msg: "角色不存在"}
	USER_ROLE_NOT_FOUND    = ResponseCode{Code: 409, Msg: "用户角色不存在"}
	USER_ACCOUNT_DUPLICATE = ResponseCode{Code: 410, Msg: "该账号已被使用"}
	EMAIL_DUPLICATE        = ResponseCode{Code: 411, Msg: "该邮箱已被使用"}
	PLATFORM_LOGGED_IN     = ResponseCode{Code: 412, Msg: "该平台已登录"}
	NOT_ALLOWED_DIR        = ResponseCode{Code: 413, Msg: "不支持的上传目录"}
	USER_BANNED            = ResponseCode{Code: 414, Msg: "该账号已被封禁"}

	// 房间相关:431-460
	ROOM_NOT_FOUND       = ResponseCode{Code: 431, Msg: "房间不存在"}
	ROOM_PERMISSION_DENY = ResponseCode{Code: 432, Msg: "权限不足，无法执行该操作"}

	// 令牌相关：600-630
	REFRESH_TOKEN_INVALID = ResponseCode{Code: 601, Msg: "无效的刷新令牌"}
	REFRESH_TOKEN_REVOKED = ResponseCode{Code: 602, Msg: "刷新令牌已撤销"}
	GENERATE_TOKEN_ERROR  = ResponseCode{Code: 603, Msg: "生成token失败"}
)
