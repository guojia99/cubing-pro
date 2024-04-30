package middleware

import (
	"fmt"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/patrickmn/go-cache"
)

const authUserKey = "authUserKey"

var checkRoleAuth *CheckAuth
var checkRoleAuthOnce sync.Once

type CheckAuth struct {
	cache *cache.Cache
	svc   *svc.Svc
}

func ClearCheckAuth() {
	if checkRoleAuth == nil {
		return
	}
	checkRoleAuth.cache.Flush()
}

func InitCheckAuth(svc *svc.Svc) {
	checkRoleAuthOnce.Do(
		func() {
			checkRoleAuth = &CheckAuth{
				cache: cache.New(time.Second*30, time.Minute),
				svc:   svc,
			}
		},
	)
}

func CheckAuthMiddlewareFunc(auth user2.Auth) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// 获取上下文权限
		user, err := GetJwtUser(ctx)
		if err != nil {
			exception.ErrAuthField.ResponseWithError(ctx, err)
			return
		}

		// 查询缓存和数据库
		key := fmt.Sprintf("%d", user.Id)
		var dbUser user2.User
		find, ok := checkRoleAuth.cache.Get(key)
		if ok {
			dbUser = find.(user2.User)
		} else {
			if err = checkRoleAuth.svc.DB.Where("id = ?", user.Id).First(&dbUser).Error; err != nil {
				exception.ErrAuthField.ResponseWithError(ctx, err)
				return
			}
		}

		// 对比权限
		if !dbUser.CheckAuth(auth) {
			exception.ErrAuthField.ResponseWithError(ctx, "权限不足")
			return
		}

		ctx.Set(authUserKey, dbUser)
		ctx.Next()
	}
}
