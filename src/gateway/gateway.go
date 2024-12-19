package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/unrolled/secure"
)

type Gateway struct {
	api *gin.Engine
	cfg svc.Config
}

func NewGateway(svc *svc.Svc) *Gateway {
	return &Gateway{
		cfg: svc.Cfg,
		api: gin.Default(),
	}
}

func (g *Gateway) Run() error {
	g.api.NoRoute(g.baseRoute())
	//g.api.Static("/", g.cfg.Gateway.StaticPath)
	g.api.Use(tlsHandler(g.cfg.Gateway.HTTPSPort, g.cfg.Gateway.HTTPSHost))
	return g.api.RunTLS(fmt.Sprintf(":%d", g.cfg.Gateway.HTTPSPort),
		g.cfg.Gateway.PEM, g.cfg.Gateway.PrivateKey)
}

func (g *Gateway) baseRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://%s:%d", g.cfg.APIConfig.Host, g.cfg.APIConfig.Port))
	proxyApi := httputil.NewSingleHostReverseProxy(api)

	return func(ctx *gin.Context) {
		if strings.Contains(ctx.Request.URL.Path, "/v3/cube-api") {
			proxyApi.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		ext := path.Ext(ctx.Request.URL.Path)
		if slices.Contains([]string{".css", ".js", ".svg", ".webp", ".woff", ".png", ".jpeg", ".jpg", ".ico"}, ext) {
			staticFilePath := filepath.Join(g.cfg.Gateway.StaticPath, ctx.Request.URL.Path)
			ctx.File(staticFilePath)
			return
		}
		ctx.Header("Cache-Control", "public, max-age=2592000")
		ctx.File(g.cfg.Gateway.IndexPath)
	}
}

func tlsHandler(port int, host string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secureMiddleware := secure.New(
			secure.Options{
				SSLRedirect: true,
				SSLHost:     host + ":" + strconv.Itoa(port),
			},
		)
		err := secureMiddleware.Process(c.Writer, c.Request)

		// If there was an error, do not continue.
		if err != nil {
			c.AbortWithStatusJSON(
				http.StatusBadGateway, gin.H{
					"error": err,
				},
			)
			return
		}

		c.Next()
	}
}
