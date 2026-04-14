package gateway

import (
	"fmt"
	"log"

	"github.com/gin-contrib/gzip"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/svc"
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
