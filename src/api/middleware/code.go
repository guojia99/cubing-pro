package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/mojocn/base64Captcha"

	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

var codeMiddleware *code
var codeOnce = sync.Once{}

const timeoutTime = time.Minute * 3

func Code() *code {
	codeOnce.Do(
		func() {
			codeMiddleware = &code{
				result: base64Captcha.NewMemoryStore(10240, timeoutTime),
			}
		},
	)
	return codeMiddleware
}

func (c *code) VerifyCaptcha(id, verifyValue string) bool {
	return c.result.Verify(id, verifyValue, true)
}

type code struct {
	result base64Captcha.Store
}

type CodeResp struct {
	Id    string    `json:"id"`
	Image string    `json:"image"`
	Ext   time.Time `json:"ext"`
}

func (c *code) CodeRouter() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		res := base64Captcha.NewCaptcha(utils.MathRandomConfig(), c.result)

		id, base64, value, err := res.Generate()
		if err != nil {
			exception.ErrInternalServer.ResponseWithError(ctx, err)
			return
		}
		fmt.Println(value)

		ctx.JSON(
			http.StatusOK, CodeResp{
				Id:    id,
				Image: base64,
				Ext:   time.Now().Add(timeoutTime - time.Second),
			},
		)
	}
}

type VerifyCodeReq struct {
	// 验证码
	VerifyId    string `query:"verifyId"`
	VerifyValue string `query:"verifyValue"`
}

func (c *code) VerifyCodeMiddlewareFn(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if !svc.Cfg.GlobalConfig.Debug {
			ctx.Next()
			return
		}

		var req VerifyCodeReq
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrVerifyCodeField.ResponseWithError(ctx, err)
			return
		}

		if ok := Code().VerifyCaptcha(req.VerifyId, req.VerifyValue); !ok {
			exception.ErrVerifyCodeField.ResponseWithError(ctx, "验证码错误")
			return
		}
	}
	// 验证验证码

}
