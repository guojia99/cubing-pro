package exception

import (
	"fmt"
	"net/http"
	"runtime"

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

func (e ErrorMsg) Error() string {
	return fmt.Sprintf("%s", e.Msg)
}

type ErrorMsg struct {
	Code     int `json:"code"`      // 业务错误码
	HttpCode int `json:"http_code"` // http 错误码

	Msg      string      `json:"message"` // 错误叠加
	ErrorMsg interface{} `json:"error"`   // 错误原始
	Data     interface{} `json:"data"`    // 补充数据

	Level string `json:"level"` // 错误级别
	Line  string `json:"line"`  // 错误行
	Ref   string `json:"ref"`   // 参考链接
}

func (e ErrorMsg) ResponseWithError(ctx *gin.Context, err interface{}) {
	if e.HttpCode == 0 {
		e.HttpCode = http.StatusBadRequest
	}
	e.Data = err
	if err != nil {
		e.ErrorMsg = err
		_, file, line, _ := runtime.Caller(1)
		e.Line = fmt.Sprintf("err at %s:%d", file, line)
	}
	ctx.AbortWithStatusJSON(e.HttpCode, e)
}

func (e ErrorMsg) ResponseWithData(ctx *gin.Context, data interface{}, err error) {
	e.Data = data
	e.ResponseWithError(ctx, err)
}
