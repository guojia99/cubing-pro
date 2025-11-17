package gateway

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strconv"
	"strings"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/unrolled/secure"
)

type Gateway struct {
	api *gin.Engine

	tNoodleApi *gin.Engine
	cfg        configs.Config
}

func NewGateway(svc *svc.Svc) *Gateway {
	return &Gateway{
		cfg: svc.Cfg,
		api: gin.Default(),
	}
}

func (g *Gateway) runTNoodleApi() {
	if g.cfg.Gateway.OutsizeTNoodlePort == 0 {
		return
	}
	log.Printf("start tNoodle api %d -> %d\n", g.cfg.Gateway.OutsizeTNoodlePort, g.cfg.Gateway.TNoodlePort)

	g.tNoodleApi = gin.Default()
	g.tNoodleApi.NoRoute(g.tNoodleRoute())
	err := g.tNoodleApi.Run(fmt.Sprintf(":%d", g.cfg.Gateway.OutsizeTNoodlePort))
	if err != nil {
		log.Fatalf("failed to start tNoodle api: %v", err)
	}
}

func (g *Gateway) tNoodleRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://localhost:%d", g.cfg.Gateway.TNoodlePort))
	proxyTNoodleApi := httputil.NewSingleHostReverseProxy(api)

	return func(ctx *gin.Context) {
		proxyTNoodleApi.ServeHTTP(ctx.Writer, ctx.Request)
	}
}

func (g *Gateway) Run() error {
	g.api.Use(gzip.Gzip(gzip.DefaultCompression, gzip.WithExcludedPaths([]string{"/v3/cube-api"})))
	g.api.NoRoute(g.baseRoute())
	//g.api.Static("/", g.cfg.Gateway.StaticPath)
	go g.runTNoodleApi()
	// 监听cubing pro api
	if g.cfg.Gateway.HTTPSPort > 0 {
		g.api.Use(tlsHandler(g.cfg.Gateway.HTTPSPort, g.cfg.Gateway.HTTPSHost))
		_ = g.api.RunTLS(fmt.Sprintf(":%d", g.cfg.Gateway.HTTPSPort),
			g.cfg.Gateway.PEM, g.cfg.Gateway.PrivateKey) // 开启443
		log.Println("http server listening on :", g.cfg.Gateway.HTTPSPort)
	}
	return g.api.Run(fmt.Sprintf(":%d", g.cfg.Gateway.HttpPort)) // 开启80
}

func (g *Gateway) baseRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://%s:%d", g.cfg.APIConfig.Host, g.cfg.APIConfig.Port))
	proxyApi := httputil.NewSingleHostReverseProxy(api)

	return func(ctx *gin.Context) {
		if strings.Contains(ctx.Request.URL.Path, "/v3/cube-api") {
			proxyApi.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		if strings.Contains(ctx.Request.URL.Path, "/v3/x-file") {
			filename := strings.ReplaceAll(ctx.Request.URL.Path, "/v3/x-file", "")
			filePath := path.Join(g.cfg.Gateway.XFile, filename)
			ctx.File(filePath)
			return
		}

		ext := path.Ext(ctx.Request.URL.Path)
		if slices.Contains([]string{".css", ".js", ".svg", ".webp", ".woff", ".png", ".jpeg", ".jpg", ".ico"}, ext) {
			staticFilePath := filepath.Join(g.cfg.Gateway.StaticPath, ctx.Request.URL.Path)
			ctx.Header("Cache-Control", "public, max-age=2592000")
			ctx.File(staticFilePath)
			return
		}

		ctx.Header("Cache-Control", "public, max-age=10")
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
