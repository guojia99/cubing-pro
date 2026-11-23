package formulaforum

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
)

type commentAttempt struct {
	userID uint
	at     time.Time
}

type commentWindow struct {
	mu       sync.Mutex
	limit    int
	window   time.Duration
	now      func() time.Time
	attempts []commentAttempt
	counts   map[uint]int
}

func CommentRateLimit(limit int, window time.Duration) gin.HandlerFunc {
	limiter := &commentWindow{limit: limit, window: window, now: time.Now, counts: make(map[uint]int)}
	return func(ctx *gin.Context) {
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}
		if !limiter.allow(user.ID) {
			exception.ErrRateLimitExceeded.ResponseWithError(ctx, "发言过于频繁，请稍后再试")
			return
		}
		ctx.Next()
	}
}

func (limiter *commentWindow) allow(userID uint) bool {
	limiter.mu.Lock()
	defer limiter.mu.Unlock()

	now := limiter.now()
	cutoff := now.Add(-limiter.window)
	for len(limiter.attempts) > 0 && !limiter.attempts[0].at.After(cutoff) {
		expired := limiter.attempts[0].userID
		limiter.attempts = limiter.attempts[1:]
		limiter.counts[expired]--
		if limiter.counts[expired] == 0 {
			delete(limiter.counts, expired)
		}
	}
	if limiter.counts[userID] >= limiter.limit {
		return false
	}
	limiter.attempts = append(limiter.attempts, commentAttempt{userID: userID, at: now})
	limiter.counts[userID]++
	return true
}
