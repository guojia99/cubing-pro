package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"sync"
	"time"
)

type RateLimitMiddlewareRequestInfo struct {
	lastRequestTime time.Time
	requestCount    int
}

func RateLimitMiddleware(limit int, duration time.Duration) gin.HandlerFunc {
	var clients sync.Map

	return func(ctx *gin.Context) {
		clientIP := ctx.ClientIP()
		now := time.Now()

		val, _ := clients.LoadOrStore(clientIP, &RateLimitMiddlewareRequestInfo{now, 0})
		info := val.(*RateLimitMiddlewareRequestInfo)

		infoMutex := &sync.Mutex{}
		infoMutex.Lock()

		defer infoMutex.Unlock()

		if now.Sub(info.lastRequestTime) > duration {
			info.requestCount = 0
			info.lastRequestTime = now
		}

		info.requestCount++

		if info.requestCount > limit {
			exception.ErrRateLimitExceeded.ResponseWithError(ctx, "请求过快")
			return
		}
		info.lastRequestTime = now
		ctx.Next()
	}
}
