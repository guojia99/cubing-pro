package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"golang.org/x/oauth2"
)

const (
	wcaProfileURL     = "https://www.worldcubeassociation.org/api/v0/me"
	wcaTokenURL       = "https://www.worldcubeassociation.org/oauth/token"
	wcaAuthorizeURL   = "https://www.worldcubeassociation.org/oauth/authorize"
	wcaStateExpireMin = 10 // state 有效期（分钟），需大于 WCA code 有效期
	wcaDefaultScopes  = "public dob email openid profile"
)

// StatePayload 用于构造 state，防 CSRF
type statePayload struct {
	Redirect string `json:"r"`
	Nonce    string `json:"n"`
}

func generateNonce() string {
	b := make([]byte, 16)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

func getWcaOAuthConfig(cfg configs.WcaAuth2, redirectURI string) *oauth2.Config {
	scopes := cfg.Auths
	if len(scopes) == 0 {
		scopes = strings.Split(wcaDefaultScopes, " ")
	}
	return &oauth2.Config{
		ClientID:     cfg.AppID,
		ClientSecret: cfg.AppSecret,
		RedirectURL:  redirectURI,
		Scopes:       scopes,
		Endpoint: oauth2.Endpoint{
			AuthURL:  wcaAuthorizeURL,
			TokenURL: wcaTokenURL,
		},
	}
}

func isValidRedirectURI(redirect, frontendBase string) bool {
	if redirect == "" || redirect == "debug" {
		return true
	}
	u, err := url.Parse(redirect)
	if err != nil {
		return false
	}
	return u.Scheme+"://"+u.Host == frontendBase || strings.HasPrefix(redirect, frontendBase+"/")
}

// redirectToWcaAuth 重定向到 WCA 授权页（用于 code 过期时重新授权）
func redirectToWcaAuth(ctx *gin.Context, svc *svc.Svc, cfg configs.WcaAuth2, redirectURI, redirect string) {
	nonce := generateNonce()
	expiresAt := time.Now().Add(wcaStateExpireMin * time.Minute)
	stateRecord := user.OAuthState{
		Nonce:     nonce,
		Redirect:  redirect,
		ExpiresAt: expiresAt,
	}
	if err := svc.DB.Create(&stateRecord).Error; err != nil {
		exception.ErrInternalServer.ResponseWithError(ctx, "failed to create auth state")
		return
	}
	stateData := statePayload{Redirect: redirect, Nonce: nonce}
	stateBytes, _ := json.Marshal(stateData)
	state := base64.RawURLEncoding.EncodeToString(stateBytes)
	oauthCfg := getWcaOAuthConfig(cfg, redirectURI)
	authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
	ctx.Redirect(http.StatusTemporaryRedirect, authURL)
}

// WcaAuthInit 发起 WCA 登录，302 跳转到 WCA 授权页
func WcaAuthInit(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := svc.Cfg.GlobalConfig.WcaAuth2
		if cfg.AppID == "" || cfg.AppSecret == "" {
			exception.ErrInternalServer.ResponseWithError(ctx, "WCA OAuth 未配置")
			return
		}

		redirect := ctx.Query("redirect")
		frontendBase := cfg.FrontendBase
		if frontendBase == "" {
			frontendBase = cfg.RedirectBase
		}
		if !isValidRedirectURI(redirect, frontendBase) {
			exception.ErrInvalidInput.ResponseWithError(ctx, "invalid redirect URI")
			return
		}

		redirectBase := cfg.RedirectBase
		if redirectBase == "" {
			redirectBase = svc.Cfg.GlobalConfig.BaseHost
		}
		callbackPath := "/v3/cube-api/auth/wca/callback"
		redirectURI := strings.TrimSuffix(redirectBase, "/") + callbackPath

		nonce := generateNonce()
		expiresAt := time.Now().Add(wcaStateExpireMin * time.Minute)
		stateRecord := user.OAuthState{
			Nonce:     nonce,
			Redirect:  redirect,
			ExpiresAt: expiresAt,
		}
		if err := svc.DB.Create(&stateRecord).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		stateData := statePayload{Redirect: redirect, Nonce: nonce}
		stateBytes, _ := json.Marshal(stateData)
		state := base64.RawURLEncoding.EncodeToString(stateBytes)

		oauthCfg := getWcaOAuthConfig(cfg, redirectURI)
		authURL := oauthCfg.AuthCodeURL(state, oauth2.AccessTypeOnline)
		ctx.Redirect(http.StatusTemporaryRedirect, authURL)
	}
}

// WcaAuthCallback WCA 回调，用 code 换 token、获取用户、生成 JWT、跳回前端或返回调试信息
func WcaAuthCallback(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		cfg := svc.Cfg.GlobalConfig.WcaAuth2
		if cfg.AppID == "" || cfg.AppSecret == "" {
			exception.ErrInternalServer.ResponseWithError(ctx, "WCA OAuth 未配置")
			return
		}

		code := ctx.Query("code")
		state := ctx.Query("state")
		if code == "" || state == "" {
			exception.ErrInvalidInput.ResponseWithError(ctx, "missing code or state")
			return
		}

		stateBytes, err := base64.RawURLEncoding.DecodeString(state)
		if err != nil {
			exception.ErrInvalidInput.ResponseWithError(ctx, "invalid state format")
			return
		}

		var payload statePayload
		if err := json.Unmarshal(stateBytes, &payload); err != nil {
			exception.ErrInvalidInput.ResponseWithError(ctx, "invalid state payload")
			return
		}

		redirectBase := cfg.RedirectBase
		if redirectBase == "" {
			redirectBase = svc.Cfg.GlobalConfig.BaseHost
		}
		callbackPath := "/v3/cube-api/auth/wca/callback"
		redirectURI := strings.TrimSuffix(redirectBase, "/") + callbackPath

		var stateRecord user.OAuthState
		if err = svc.DB.Where("nonce = ?", payload.Nonce).First(&stateRecord).Error; err != nil {
			// state 不存在或已过期（表不存在、服务器重启、跨实例等），重定向到 WCA 重新授权
			redirectToWcaAuth(ctx, svc, cfg, redirectURI, payload.Redirect)
			return
		}
		if time.Now().After(stateRecord.ExpiresAt) {
			_ = svc.DB.Delete(&stateRecord)
			redirectToWcaAuth(ctx, svc, cfg, redirectURI, payload.Redirect)
			return
		}
		_ = svc.DB.Delete(&stateRecord)

		oauthCfg := getWcaOAuthConfig(cfg, redirectURI)
		token, err := oauthCfg.Exchange(context.Background(), code)
		if err != nil {
			// code 过期或无效，重定向到 WCA 重新授权
			redirectToWcaAuth(ctx, svc, cfg, redirectURI, payload.Redirect)
			return
		}

		client := oauthCfg.Client(context.Background(), token)
		resp, err := client.Get(wcaProfileURL)
		if err != nil || resp.StatusCode != http.StatusOK {
			exception.ErrInternalServer.ResponseWithError(ctx, "failed to fetch WCA user")
			return
		}
		defer resp.Body.Close()

		var wcaResp struct {
			Me struct {
				ID        int    `json:"id"`
				WCAID     string `json:"wca_id"`
				Name      string `json:"name"`
				Email     string `json:"email"`
				Country   string `json:"country_iso2"`
				Gender    string `json:"gender"`
				Birthdate string `json:"birthdate"`
				Avatar    struct {
					URL string `json:"url"`
				} `json:"avatar"`
			} `json:"me"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&wcaResp); err != nil {
			exception.ErrInternalServer.ResponseWithError(ctx, "failed to parse WCA user")
			return
		}

		wcaUser := wcaResp.Me
		if wcaUser.WCAID == "" {
			exception.ErrInternalServer.ResponseWithError(ctx, "WCA user has no wca_id")
			return
		}

		avatarURL := wcaUser.Avatar.URL

		// 查找或创建用户
		var dbUser user.User
		err = svc.DB.Where("wca_id = ?", wcaUser.WCAID).First(&dbUser).Error
		if err != nil {
			// 新用户，创建
			now := time.Now()
			dbUser = user.User{
				WcaID:             wcaUser.WCAID,
				Name:              wcaUser.Name,
				Email:             wcaUser.Email,
				Avatar:            avatarURL,
				Nationality:       wcaUser.Country,
				Auth:              user.AuthPlayer,
				WcaLoginAt:        &now,
				WcaAccessToken:    token.AccessToken,
				WcaTokenExpiresAt: nil,
			}
			if token.Expiry.IsZero() {
				exp := time.Now().Add(2 * time.Hour)
				dbUser.WcaTokenExpiresAt = &exp
			} else {
				dbUser.WcaTokenExpiresAt = &token.Expiry
			}
			if err = svc.DB.Create(&dbUser).Error; err != nil {
				exception.ErrDatabase.ResponseWithError(ctx, err)
				return
			}
		} else {
			// 已存在，关联并更新
			now := time.Now()
			dbUser.WcaLoginAt = &now
			dbUser.WcaAccessToken = token.AccessToken
			if token.Expiry.IsZero() {
				exp := time.Now().Add(2 * time.Hour)
				dbUser.WcaTokenExpiresAt = &exp
			} else {
				dbUser.WcaTokenExpiresAt = &token.Expiry
			}
			if dbUser.Name == "" && wcaUser.Name != "" {
				dbUser.Name = wcaUser.Name
			}
			if dbUser.Email == "" && wcaUser.Email != "" {
				dbUser.Email = wcaUser.Email
			}
			if dbUser.Avatar == "" && avatarURL != "" {
				dbUser.Avatar = avatarURL
			}
			if err = svc.DB.Save(&dbUser).Error; err != nil {
				exception.ErrDatabase.ResponseWithError(ctx, err)
				return
			}
		}

		claims := middleware.JwtMapClaims{
			Id:           dbUser.ID,
			Auth:         dbUser.Auth,
			Name:         dbUser.Name,
			EnName:       dbUser.EnName,
			LoginID:      dbUser.LoginID,
			CubeID:       dbUser.CubeID,
			WcaID:        dbUser.WcaID,
			DelegateName: dbUser.DelegateName,
		}

		tokenStr, _, err := middleware.JWT().TokenGenerator(claims)
		if err != nil {
			exception.ErrInternalServer.ResponseWithError(ctx, "failed to generate token")
			return
		}

		// 调试模式：redirect 为空或 "debug" 时返回 JSON
		if payload.Redirect == "" || payload.Redirect == "debug" {
			ctx.JSON(http.StatusOK, gin.H{
				"token": tokenStr,
				"user":  claims,
				"wca": gin.H{
					"wca_id": wcaUser.WCAID,
					"name":   wcaUser.Name,
					"email":  wcaUser.Email,
					"avatar": avatarURL,
				},
			})
			return
		}

		// 正常模式：跳转到前端并附带 token
		target := payload.Redirect
		frontendBase := cfg.FrontendBase
		if frontendBase == "" {
			frontendBase = cfg.RedirectBase
		}
		if target == "" {
			target = frontendBase
		}
		finalURL := target
		if strings.Contains(target, "?") {
			finalURL += "&token=" + url.QueryEscape(tokenStr)
		} else {
			finalURL += "?token=" + url.QueryEscape(tokenStr)
		}
		ctx.Redirect(http.StatusFound, finalURL)
	}
}

// WcaAuthMe 调试用：需携带 token 获取当前用户信息（用于直连后端测试）
func WcaAuthMe(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authUser, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return // GetAuthUser 已处理错误响应
		}
		ctx.JSON(http.StatusOK, gin.H{
			"user": authUser,
		})
	}
}
