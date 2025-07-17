package common

// ResponseCode 定义响应码和消息的组合
type ResponseCode struct {
	Code int
	Msg  string
}

// 预定义响应状态
var (
	SUCCESS                = ResponseCode{Code: 200, Msg: "操作成功"}
	ERROR                  = ResponseCode{Code: 500, Msg: "操作失败"}
	INVALID_PARAMS         = ResponseCode{Code: 400, Msg: "请求参数错误"}
	USER_ACCOUNT_EXISTS    = ResponseCode{Code: 401, Msg: "用户账号已存在"}
	PASSWORD_INVALID       = ResponseCode{Code: 402, Msg: "密码错误"}
	EMAIL_INVALID          = ResponseCode{Code: 403, Msg: "邮箱错误"}
	REFRESH_TOKEN_INVALID  = ResponseCode{Code: 404, Msg: "无效的刷新令牌"}
	REFRESH_TOKEN_REVOKED  = ResponseCode{Code: 405, Msg: "刷新令牌已撤销"}
	GENERATE_TOKEN_ERROR   = ResponseCode{Code: 406, Msg: "生成token失败"}
	USER_ID_NOT_FOUND      = ResponseCode{Code: 407, Msg: "用户id不存在"}
	USER_ACCOUNT_NOT_FOUND = ResponseCode{Code: 408, Msg: "用户账号不存在"}
	USER_NOT_FOUND         = ResponseCode{Code: 409, Msg: "用户不存在"}
)
