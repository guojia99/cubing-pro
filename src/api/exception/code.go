package exception

import "net/http"

const (
	H400 = http.StatusBadRequest
	H401 = http.StatusUnauthorized
	H403 = http.StatusForbidden
	H404 = http.StatusNotFound
	H409 = http.StatusConflict
	H422 = http.StatusUnprocessableEntity
	H429 = http.StatusTooManyRequests
	H500 = http.StatusInternalServerError
	H503 = http.StatusServiceUnavailable
	H504 = http.StatusGatewayTimeout
)

// common errors 10001 ~ 10499
var (
	ErrRequestBinding     = NewErrorMsg(H400, 10001, "错误参数", "", "")
	ErrRequestHeaderField = NewErrorMsg(H400, 10002, "错误的头参数", "", "")
	ErrJwtField           = NewErrorMsg(H401, 10003, "Jwt解析错误", "", "")
	ErrAuthField          = NewErrorMsg(H401, 10004, "无权限", "", "")
	ErrGetData            = NewErrorMsg(H404, 10005, "数据不存在", "", "")
	ErrUserNotFound       = NewErrorMsg(H401, 10006, "用户不存在", "", "")
	ErrPasswordField      = NewErrorMsg(H401, 10007, "密码错误", "", "")
	ErrInvalidInput       = NewErrorMsg(H400, 10008, "无效输入", "", "")
	ErrDatabase           = NewErrorMsg(H500, 10009, "数据库错误", "", "")
	ErrInternalServer     = NewErrorMsg(H500, 10010, "服务器内部错误", "", "")
	ErrUnauthorized       = NewErrorMsg(H401, 10011, "未经授权的访问", "", "")
	ErrForbidden          = NewErrorMsg(H403, 10012, "禁止访问", "", "")
	ErrResourceNotFound   = NewErrorMsg(H404, 10013, "资源不存在", "", "")
	ErrValidationFailed   = NewErrorMsg(H422, 10014, "验证失败", "", "")
	ErrTokenExpired       = NewErrorMsg(H401, 10015, "令牌已过期", "", "")
	ErrServerBusy         = NewErrorMsg(H503, 10018, "服务器繁忙", "", "")
	ErrServiceUnavailable = NewErrorMsg(H503, 10019, "服务不可用", "", "")
	ErrRateLimitExceeded  = NewErrorMsg(H429, 10020, "超过速率限制", "", "")
	ErrInvalidCredentials = NewErrorMsg(H401, 10021, "凭据无效", "", "")
	ErrExpiredSession     = NewErrorMsg(H401, 10022, "会话已过期", "", "")
	ErrResourceConflict   = NewErrorMsg(H409, 10025, "资源冲突", "", "")
	ErrUnavailableService = NewErrorMsg(H503, 10026, "服务不可用", "", "")
	ErrGatewayTimeout     = NewErrorMsg(H504, 10027, "网关超时", "", "")
	ErrInvalidTokenFormat = NewErrorMsg(H401, 10028, "无效令牌格式", "", "")
	ErrResourceForbidden  = NewErrorMsg(H403, 10033, "资源禁止访问", "", "")
	ErrVerifyCodeField    = NewErrorMsg(H401, 10032, "验证码错误", "", "")
)

// auth errors 10500 ~ 10999
var (
	ErrRegisterField = NewErrorMsg(H401, 10500, "注册错误", "", "")
)

// system errors 11000 ~ 12000
// comps errors 12001 ~ 13000
// result errors 13001 ~ 14000

var (
	ErrResultScoreFormatField = NewErrorMsg(H400, 13001, "成绩格式错误", "", "")
)
