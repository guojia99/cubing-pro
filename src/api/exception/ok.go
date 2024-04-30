package exception

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type OkMsg struct {
	Code string      `json:"code"`
	Data interface{} `json:"data"`
	Msg  string      `json:"msg"`
}

func ResponseOK(ctx *gin.Context, data interface{}) {
	ctx.JSON(
		http.StatusOK, OkMsg{
			Code: "200",
			Data: data,
			Msg:  "ok",
		},
	)
}
