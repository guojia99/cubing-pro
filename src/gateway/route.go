package gateway

import (
	"fmt"
	"net/http/httputil"
	"net/url"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
)

func localProxyTargetHost(p configs.LocalProxyConfig) string {
	if h := strings.TrimSpace(p.TargetHost); h != "" {
		return h
	}
	return "127.0.0.1"
}

func localProxyHostKeys(p configs.LocalProxyConfig) []string {
	if len(p.Hosts) > 0 {
		out := make([]string, 0, len(p.Hosts))
		for _, h := range p.Hosts {
			h = strings.ToLower(strings.TrimSpace(h))
			if h != "" {
				out = append(out, h)
			}
		}
		return out
	}
	h := strings.ToLower(strings.TrimSpace(p.Host))
	if h == "" {
		return nil
	}
	return []string{h}
}

func buildLocalProxyMap(proxies []configs.LocalProxyConfig) map[string]*httputil.ReverseProxy {
	m := make(map[string]*httputil.ReverseProxy)
	for _, p := range proxies {
		if p.Port <= 0 {
			continue
		}
		u, err := url.Parse(fmt.Sprintf("http://%s:%d", localProxyTargetHost(p), p.Port))
		if err != nil {
			continue
		}
		px := httputil.NewSingleHostReverseProxy(u)
		configureReverseProxy(px)
		for _, key := range localProxyHostKeys(p) {
			m[key] = px
		}
	}
	return m
}

func (g *Gateway) baseRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://%s:%d", g.cfg.APIConfig.Host, g.cfg.APIConfig.Port))
	proxyApi := httputil.NewSingleHostReverseProxy(api)
	configureReverseProxy(proxyApi)

	blddbApi, _ := url.Parse(fmt.Sprintf("http://localhost:%d", g.cfg.Gateway.BldDBPort))
	bldDbProxy := httputil.NewSingleHostReverseProxy(blddbApi)
	configureReverseProxy(bldDbProxy)

	staticSites := buildStaticSiteMap(g.cfg.Gateway.StaticSites)
	localProxies := buildLocalProxyMap(g.cfg.Gateway.LocalProxies)

	return func(ctx *gin.Context) {
		host := normalizeRequestHost(ctx.Request.Host)

		// blddb
		if host == "blddb.cubing.pro" {
			bldDbProxy.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		if px, ok := localProxies[host]; ok {
			px.ServeHTTP(ctx.Writer, ctx.Request)
			return
		}

		if site, ok := staticSites[host]; ok {
			serveStaticSite(ctx, site)
			return
		}
		//
		//if strings.Contains(ctx.Request.URL.Path, "/blddb") {
		//	bldDbProxy.ServeHTTP(ctx.Writer, ctx.Request)
		//	return
		//}

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

		serveDefaultStatic(ctx, g.cfg.Gateway)
	}
}
