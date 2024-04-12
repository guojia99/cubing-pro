package auth

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mojocn/base64Captcha"

	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/svc"
)

var codeResult = base64Captcha.NewMemoryStore(10240, time.Minute*1)

type CodeReq struct {
}

type CodeResp struct {
	Id    string `json:"id"`
	Image string `json:"image"`
}

func Code(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := base64Captcha.NewCaptcha(utils.MathRandomConfig(), codeResult)

		id, base64, _, err := c.Generate()
		if err != nil {
			exception.ErrInternalServer.ResponseWithError(ctx, err)
			return
		}

		ctx.JSON(
			http.StatusOK, CodeResp{
				Id:    id,
				Image: base64,
			},
		)
	}
}

func verifyCaptcha(id, verifyValue string) bool {
	return codeResult.Verify(id, verifyValue, true)
}
