package gateway

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/unrolled/secure"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strconv"
	"strings"
)

type Gateway struct {
	api *gin.Engine
	cfg svc.Config
}

func NewGateway(cfg svc.Config) *Gateway {
	return &Gateway{
		cfg: cfg,
		api: gin.Default(),
	}
}

func (g *Gateway) Run() error {
	g.api.Static("/dist", g.cfg.Gateway.StaticPath)
	g.api.NoRoute(g.baseRoute())

	if g.cfg.Gateway.PEM != "" && g.cfg.Gateway.PrivateKey != "" {
		g.api.Use(tlsHandler(g.cfg.Gateway.HTTPSPort, g.cfg.Gateway.HTTPSHost))
		go g.api.RunTLS(fmt.Sprintf(":%d", g.cfg.Gateway.HTTPSPort),
			g.cfg.Gateway.PEM, g.cfg.Gateway.PrivateKey)
	}
	return g.api.Run(fmt.Sprintf(":%d", g.cfg.Gateway.HttpPort))
}

func (g *Gateway) baseRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://%s:%d", g.cfg.APIConfig.Host, g.cfg.APIConfig.Port))
	proxyApi := httputil.NewSingleHostReverseProxy(api)

	return func(ctx *gin.Context) {
		if strings.Contains(ctx.Request.URL.Path, "/v3/cube-api") {
			proxyApi.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}
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
