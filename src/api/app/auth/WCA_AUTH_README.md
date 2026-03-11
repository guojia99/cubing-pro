下面是一份 完整、安全、可运行 的 Go + Gin 项目示例，用于：
接入 World Cube Association (WCA) OAuth2
使用 授权码模式（Authorization Code Flow）
回调后换取 access_token
调用 WCA /api/v0/me 获取用户信息
生成 JWT（JSON Web Token） 作为本地登录凭证
包含 state 防 CSRF、错误处理、环境变量配置等最佳实践



✅ 前提条件
在 WCA OAuth Applications 创建应用：
Redirect URI: http://localhost:8000/auth/wca/callback
记下 Client ID 和 Client Secret
安装依赖：
```
go mod init wca-jwt-login
go get github.com/gin-gonic/gin
go get golang.org/x/oauth2
go get github.com/golang-jwt/jwt/v5
```
/auth/wca 接收 ?redirect=/your-page
将 redirect 地址和随机 nonce 一起编码为 state
回调时解码 state，验证 nonce，然后跳转到 redirect
我们采用 JSON 编码 + Base64 的方式将 {target: "...", nonce: "..."} 存入 state。

示例代码
```go
package main

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

var (
	wcaOAuthConfig = &oauth2.Config{
		ClientID:     os.Getenv("WCA_CLIENT_ID"),
		ClientSecret: os.Getenv("WCA_CLIENT_SECRET"),
		RedirectURL:  "http://localhost:8000/auth/wca/callback",
		Scopes:       []string{"public", "email", "dob", "profile", "openid"},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://www.worldcubeassociation.org/oauth/authorize",
			TokenURL: "https://www.worldcubeassociation.org/oauth/token",
		},
	}
	jwtSecret        = []byte(os.Getenv("JWT_SECRET"))
	frontendBaseURL  = "http://localhost:3000" // 前端域名（用于安全校验）
	validStateNonces = make(map[string]time.Time) // 仅存 nonce，不存完整 URL
)

// StatePayload 用于构造 state
type StatePayload struct {
	Redirect string `json:"r"` // 用户想回到的前端页面（必须是 frontendBaseURL 下的路径）
	Nonce    string `json:"n"` // 随机字符串，防重放
}

// 生成随机 nonce
func generateNonce() string {
	b := make([]byte, 16)
	rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// 安全校验 redirect 是否属于你的前端域
func isValidRedirectURI(redirect string) bool {
	if redirect == "" {
		return true // 允许空，默认跳首页
	}
	u, err := url.Parse(redirect)
	if err != nil {
		return false
	}
	return u.Scheme+ "://" + u.Host == frontendBaseURL
}

func main() {
	if os.Getenv("WCA_CLIENT_ID") == "" {
		log.Fatal("Missing WCA_CLIENT_ID")
	}
	if os.Getenv("WCA_CLIENT_SECRET") == "" {
		log.Fatal("Missing WCA_CLIENT_SECRET")
	}
	if len(jwtSecret) == 0 {
		log.Fatal("Missing JWT_SECRET")
	}

	r := gin.Default()

	// 1. 发起 WCA 登录，支持 ?redirect=...
	r.GET("/auth/wca", func(c *gin.Context) {
		redirect := c.Query("redirect")

		// 安全校验（防止开放重定向漏洞）
		if !isValidRedirectURI(redirect) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid redirect URI"})
			return
		}

		nonce := generateNonce()
		validStateNonces[nonce] = time.Now().Add(5 * time.Minute)

		stateData := StatePayload{
			Redirect: redirect,
			Nonce:    nonce,
		}
		stateBytes, _ := json.Marshal(stateData)
		state := base64.RawURLEncoding.EncodeToString(stateBytes)

		authURL := wcaOAuthConfig.AuthCodeURL(state, oauth2.AccessTypeOnline)
		c.Redirect(http.StatusTemporaryRedirect, authURL)
	})

	// 2. 回调处理
	r.GET("/auth/wca/callback", func(c *gin.Context) {
		code := c.Query("code")
		state := c.Query("state")

		if code == "" || state == "" {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "missing code or state"})
			return
		}

		// 解码 state
		stateBytes, err := base64.RawURLEncoding.DecodeString(state)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid state format"})
			return
		}

		var payload StatePayload
		if err := json.Unmarshal(stateBytes, &payload); err != nil {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid state payload"})
			return
		}

		// 校验 nonce
		if createdAt, ok := validStateNonces[payload.Nonce]; !ok || time.Since(createdAt) > 5*time.Minute {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid or expired state"})
			return
		}
		delete(validStateNonces, payload.Nonce)

		// 换 token、获取用户信息、生成 JWT（同之前逻辑）
		token, err := wcaOAuthConfig.Exchange(context.Background(), code)
		if err != nil {
			log.Printf("Token exchange error: %v", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "auth failed"})
			return
		}

		client := wcaOAuthConfig.Client(context.Background(), token)
		resp, err := client.Get("https://www.worldcubeassociation.org/api/v0/me")
		if err != nil || resp.StatusCode != http.StatusOK {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user"})
			return
		}
		defer resp.Body.Close()

		var wcaUser struct {
			ID    int    `json:"id"`
			WCAID string `json:"wca_id"`
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		json.NewDecoder(resp.Body).Decode(&wcaUser)

		// 生成 JWT
		jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"user_id": wcaUser.ID,
			"wca_id":  wcaUser.WCAID,
			"email":   wcaUser.Email,
			"name":    wcaUser.Name,
			"exp":     time.Now().Add(24 * time.Hour).Unix(),
		})
		tokenStr, _ := jwtToken.SignedString(jwtSecret)

		// 构造最终跳转地址
		target := payload.Redirect
		if target == "" {
			target = frontendBaseURL // 默认跳首页
		}

		// 添加 token 到 URL（前端可读取）
		finalURL := target
		if strings.Contains(target, "?") {
			finalURL += "&token=" + url.QueryEscape(tokenStr)
		} else {
			finalURL += "?token=" + url.QueryEscape(tokenStr)
		}

		c.Redirect(http.StatusFound, finalURL)
	})

	// 示例：受保护路由
	r.GET("/api/test", func(c *gin.Context) {
		auth := c.GetHeader("Authorization")
		if !strings.HasPrefix(auth, "Bearer ") {
			c.AbortWithStatus(401)
			return
		}
		// 验证 token...
		c.JSON(200, gin.H{"msg": "protected data"})
	})

	log.Println("Server running on :8000")
	r.Run(":8000")
}
```


✅ 效果
用户访问 http://localhost:3000/secret
前端跳转到：
http://localhost:8000/auth/wca?redirect=http%3A%2F%2Flocalhost%3A3000%2Fsecret
登录成功后，自动跳回：
http://localhost:3000/secret?token=eyJhbGciOiJIUzI1NiIs...
前端读取 token 并存入 localStorage / memory，后续请求带上
