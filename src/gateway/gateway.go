package gateway

import (
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"

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

	tlsReady := g.cfg.Gateway.HTTPSPort > 0 &&
		g.cfg.Gateway.PEM != "" &&
		g.cfg.Gateway.PrivateKey != ""

	if tlsReady {
		go func() {
			addr := fmt.Sprintf(":%d", g.cfg.Gateway.HTTPSPort)
			log.Printf("gateway HTTPS listening on %s", addr)
			if err := g.api.RunTLS(addr, g.cfg.Gateway.PEM, g.cfg.Gateway.PrivateKey); err != nil {
				log.Fatalf("gateway RunTLS: %v", err)
			}
		}()

		httpPort := g.cfg.Gateway.HttpPort
		if httpPort == 0 {
			httpPort = 80
		}
		addr := fmt.Sprintf(":%d", httpPort)
		log.Printf("gateway HTTP on %s redirects to HTTPS :%d", addr, g.cfg.Gateway.HTTPSPort)
		return http.ListenAndServe(addr, g.httpToHTTPSRedirect())
	}

	httpPort := g.cfg.Gateway.HttpPort
	if httpPort == 0 {
		httpPort = 80
	}
	return g.api.Run(fmt.Sprintf(":%d", httpPort))
}

// httpToHTTPSRedirect 仅用于明文 HTTP 端口：301 到 HTTPS（TLS 由另一监听负责）。
func (g *Gateway) httpToHTTPSRedirect() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Forwarded-Proto") == "https" {
			g.api.ServeHTTP(w, r)
			return
		}
		if strings.HasPrefix(r.URL.Path, "/.well-known/acme-challenge/") {
			g.api.ServeHTTP(w, r)
			return
		}

		host := httpsRedirectHost(r, g.cfg.Gateway)
		target := buildHTTPSURL(host, g.cfg.Gateway.HTTPSPort, r.URL.RequestURI())
		http.Redirect(w, r, target, http.StatusMovedPermanently)
	})
}

func stripHostPort(host string) string {
	h, _, err := net.SplitHostPort(host)
	if err != nil {
		return host
	}
	return h
}

func httpsRedirectHost(r *http.Request, cfg configs.GatewayConfig) string {
	if cfg.HTTPSHost != "" {
		return cfg.HTTPSHost
	}
	return stripHostPort(r.Host)
}

func buildHTTPSURL(host string, httpsPort int, requestURI string) string {
	if httpsPort == 443 || httpsPort == 0 {
		return "https://" + host + requestURI
	}
	return fmt.Sprintf("https://%s:%d%s", host, httpsPort, requestURI)
}
