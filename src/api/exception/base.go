package exception

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewErrorMsg(HttpCode, Code int, Msg, Level, Ref string) ErrorMsg {
	return ErrorMsg{
		Code:     Code,
		HttpCode: HttpCode,
		Msg:      Msg,
		Level:    Level,
		Ref:      Ref,
	}
}

type ErrorMsg struct {
	Code     int         `json:"code"`      // 业务错误码
	HttpCode int         `json:"http_code"` // http 错误码
	Msg      string      `json:"message"`   // 错误叠加
	Data     interface{} `json:"data"`      // 补充数据
	Level    string      `json:"level"`     // 错误级别
	Ref      string      `json:"ref"`       // 参考链接
}

func (e ErrorMsg) ResponseWithError(ctx *gin.Context, err error) {
	if e.HttpCode == 0 {
		e.HttpCode = http.StatusBadRequest
	}
	if err != nil {
		e.Msg = fmt.Sprintf("%s: %s", e.Msg, err)
	}
	ctx.AbortWithStatusJSON(e.HttpCode, e)
}

func (e ErrorMsg) ResponseWithData(ctx *gin.Context, data interface{}, err error) {
	e.Data = data
	e.ResponseWithError(ctx, err)
}
