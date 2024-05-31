package middleware

import (
	"net"
	"sync"
	"time"

	jwt "github.com/appleboy/gin-jwt/v2"
	"github.com/gin-gonic/gin"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/patrickmn/go-cache"

	"github.com/guojia99/cubing-pro/src/api/exception"
	user2 "github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

const IdentityKey = "UserDetail"

var jwtOnce = sync.Once{}
var jwtMiddleware = &Jwt{}

type Jwt struct {
	*jwt.GinJWTMiddleware
	cache *cache.Cache
}

func JWT() *Jwt { return jwtMiddleware }

func InitJWT(svc *svc.Svc) *Jwt {
	jwtOnce.Do(
		func() {
			jwtMiddleware.GinJWTMiddleware, _ = jwt.New(
				&jwt.GinJWTMiddleware{
					Realm:           "cubing-pro",
					Key:             []byte("cubing-pro"),
					Timeout:         time.Hour * 3 * 24,
					MaxRefresh:      time.Hour * 3 * 24,
					IdentityKey:     IdentityKey,
					SendCookie:      true,
					PayloadFunc:     payloadFunc(svc),
					IdentityHandler: identityHandler(svc),
					Authorizator:    authorization(svc),
					Authenticator:   authenticator(svc),
					TokenLookup:     "header: Authorization, query: token, cookie: jwt",
					TokenHeadName:   "Bearer",
					TimeFunc:        time.Now,
				},
			)
			_ = jwtMiddleware.MiddlewareInit()
			jwtMiddleware.cache = cache.New(time.Minute, time.Minute)
		},
	)
	return jwtMiddleware
}

func payloadFunc(svc *svc.Svc) func(data interface{}) jwt.MapClaims {
	return func(data interface{}) jwt.MapClaims {
		return jwt.MapClaims{
			IdentityKey: data,
		}
	}
}

func identityHandler(svc *svc.Svc) func(c *gin.Context) interface{} {
	return func(ctx *gin.Context) interface{} {
		//fmt.Println("-----------==================")
		//fmt.Println(GetJwtUser(ctx))
		//fmt.Println(jwt.GetToken(ctx))

		return jwt.ExtractClaims(ctx)[IdentityKey]
	}
}

func authorization(svc *svc.Svc) func(data interface{}, c *gin.Context) bool {
	return func(data interface{}, ctx *gin.Context) bool {
		// todo token拦截
		//fmt.Println("==================")
		//fmt.Println(GetJwtUser(ctx))
		return true
	}
}

type LoginRequest struct {
	// 登录信息
	LoginID   string `json:"loginID"`  // 登录用的ID, email, phone, QQ等,
	Password  string `json:"password"` // 密码（加密后）
	TimeStamp int64  `json:"timestamp"`
	// 验证码
	VerifyId    string `json:"verifyId"`
	VerifyValue string `json:"verifyValue"`
}

func authenticator(svc *svc.Svc) func(ctx *gin.Context) (interface{}, error) {
	return func(ctx *gin.Context) (interface{}, error) {
		var req LoginRequest
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return nil, err
		}
		// 验证验证码
		if !svc.Cfg.GlobalConfig.Debug {
			if ok := Code().VerifyCaptcha(req.VerifyId, req.VerifyValue); !ok {
				return nil, exception.ErrVerifyCodeField
			}
		}

		// 解析密码
		password, err := utils.DePwdCode(req.Password, req.TimeStamp)
		if err != nil {
			return nil, exception.ErrRequestBinding
		}

		// 查询用户
		var user user2.User
		sql := svc.DB.Where("login_id = ?", req.LoginID)
		if utils.IsEmailValid(req.LoginID) {
			sql = svc.DB.Where("email = ?", req.LoginID)
		}
		if utils.IsPhoneNumberValid(req.LoginID) {
			sql = svc.DB.Where("phone = ?", req.LoginID)
		}
		if err = sql.First(&user).Error; err != nil {
			return nil, exception.ErrUserNotFound
		}

		// 对比密码是否正确
		if err = user.CheckPassword(password); err != nil {
			svc.DB.Save(user)
			return nil, err
		}

		user.LoginIp = net.ParseIP(ctx.ClientIP())
		user.Online = 1
		svc.DB.Save(user)

		return JwtMapClaims{
			Id:           user.ID,
			Auth:         user.Auth,
			Name:         user.Name,
			EnName:       user.EnName,
			LoginID:      user.LoginID,
			CubeID:       user.CubeID,
			WcaID:        user.WcaID,
			DelegateName: user.DelegateName,
		}, nil
	}
}

type JwtMapClaims struct {
	Id           uint       `json:"id"`
	Auth         user2.Auth `json:"auth"`
	Name         string     `json:"name"`
	EnName       string     `json:"enName"`
	LoginID      string     `json:"loginID"`
	CubeID       string     `json:"cubeID"`
	WcaID        string     `json:"wcaID"`
	DelegateName string     `json:"delegateName"`
}
