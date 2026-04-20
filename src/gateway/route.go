package gateway

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
)

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
		u, err := url.Parse(fmt.Sprintf("http://127.0.0.1:%d", p.Port))
		if err != nil {
			continue
		}
		px := httputil.NewSingleHostReverseProxy(u)
		for _, key := range localProxyHostKeys(p) {
			m[key] = px
		}
	}
	return m
}

func (g *Gateway) baseRoute() gin.HandlerFunc {
	api, _ := url.Parse(fmt.Sprintf("http://%s:%d", g.cfg.APIConfig.Host, g.cfg.APIConfig.Port))
	proxyApi := httputil.NewSingleHostReverseProxy(api)

	blddbApi, _ := url.Parse(fmt.Sprintf("http://localhost:%d", g.cfg.Gateway.BldDBPort))
	bldDbProxy := httputil.NewSingleHostReverseProxy(blddbApi)
	// 反向代理 如果是 HTML 响应，且未指定 charset，则强制加上 utf-8
	bldDbProxy.ModifyResponse = func(resp *http.Response) error {
		contentType := resp.Header.Get("Content-Type")
		if strings.HasPrefix(contentType, "text/html") && !strings.Contains(contentType, "charset") {
			resp.Header.Set("Content-Type", "text/html; charset=utf-8")
		}
		return nil
	}

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

		ctx.Header("Cache-Control", "public, max-age=60")
		ext := path.Ext(ctx.Request.URL.Path)
		if slices.Contains(getStaticFileExts(g.cfg.Gateway), ext) {
			staticFilePath := filepath.Join(g.cfg.Gateway.StaticPath, ctx.Request.URL.Path)
			ctx.File(staticFilePath)
			return
		}
		ctx.File(g.cfg.Gateway.IndexPath)
	}
}
