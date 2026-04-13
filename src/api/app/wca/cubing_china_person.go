package wca

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

// 粗饼出站：全进程串行 + 最小间隔（仅 API 层，crawler 包无锁）
var (
	cubingPersonOutboundMu sync.Mutex
	lastCubingOutboundAt   time.Time
)

const minCubingOutboundInterval = 1000 * time.Millisecond

func throttleBeforeCubingOutbound() {
	elapsed := time.Since(lastCubingOutboundAt)
	if elapsed < minCubingOutboundInterval {
		time.Sleep(minCubingOutboundInterval - elapsed)
	}
	lastCubingOutboundAt = time.Now()
}

// CubingChinaPerson 由服务端抓取粗饼选手主页（/results/person/:wcaID），返回结构化 JSON。
// 本 Handler 内全局互斥 + 节流；路由层另有 IP 限流。crawler.FetchPersonPage 无锁、内部 recover。
// HTTP 200 时 body 为 exception.ResponseOK 包装；仅参数非法时 HTTP 400。
func CubingChinaPerson(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		_ = svc
		wcaID := ctx.Param("wcaID")
		if !cubing.ValidateWcaIDFormat(strings.TrimSpace(strings.ToUpper(wcaID))) {
			exception.ErrInvalidInput.ResponseWithError(ctx, "WCA ID 格式无效（应为 10 位：4 数字 + 4 大写字母 + 2 数字）")
			return
		}

		cubingPersonOutboundMu.Lock()
		defer cubingPersonOutboundMu.Unlock()
		throttleBeforeCubingOutbound()

		cctx, cancel := context.WithTimeout(ctx.Request.Context(), 8*time.Second)
		defer cancel()

		res := cubing.FetchPersonPage(cctx, wcaID)
		if res.Code == cubing.PersonCodeInvalidWcaID {
			exception.ErrInvalidInput.ResponseWithError(ctx, res.Message)
			return
		}
		exception.ResponseOK(ctx, res)
	}
}
