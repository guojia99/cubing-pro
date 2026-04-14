package gateway

import (
	"net"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
)

func normalizeRequestHost(h string) string {
	host, _, err := net.SplitHostPort(h)
	if err != nil {
		host = h
	}
	return strings.ToLower(strings.TrimSpace(host))
}

func staticSiteHostKeys(s configs.StaticSiteConfig) []string {
	if len(s.Hosts) > 0 {
		out := make([]string, 0, len(s.Hosts))
		for _, h := range s.Hosts {
			h = strings.ToLower(strings.TrimSpace(h))
			if h != "" {
				out = append(out, h)
			}
		}
		return out
	}
	h := strings.ToLower(strings.TrimSpace(s.Host))
	if h == "" {
		return nil
	}
	return []string{h}
}

func buildStaticSiteMap(sites []configs.StaticSiteConfig) map[string]configs.StaticSiteConfig {
	m := make(map[string]configs.StaticSiteConfig)
	for _, s := range sites {
		if strings.TrimSpace(s.Root) == "" {
			continue
		}
		for _, key := range staticSiteHostKeys(s) {
			m[key] = s
		}
	}
	return m
}

func siteIndexName(s configs.StaticSiteConfig) string {
	raw := strings.TrimSpace(s.Index)
	if raw == "" {
		return "index.html"
	}
	raw = strings.TrimPrefix(raw, "/")
	return filepath.Clean(filepath.FromSlash(raw))
}

// safeJoinSiteRoot 将 URL Path 安全映射到 Root 下的绝对路径；非法穿越返回 ("", false)
func safeJoinSiteRoot(root, urlPath string) (string, bool) {
	root = filepath.Clean(root)
	up := urlPath
	if up == "" {
		up = "/"
	}
	p := strings.TrimPrefix(path.Clean("/"+up), "/")
	full := filepath.Join(root, filepath.FromSlash(p))
	full = filepath.Clean(full)
	rel, err := filepath.Rel(root, full)
	if err != nil || strings.HasPrefix(rel, "..") {
		return "", false
	}
	return full, true
}

func serveStaticSite(ctx *gin.Context, site configs.StaticSiteConfig) {
	root := filepath.Clean(site.Root)
	info, err := os.Stat(root)
	if err != nil || !info.IsDir() {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	cc := strings.TrimSpace(site.CacheControl)
	if cc == "" {
		cc = "public, max-age=60"
	}
	ctx.Header("Cache-Control", cc)

	indexRel := siteIndexName(site)
	indexPath := filepath.Join(root, indexRel)
	indexBase := filepath.Base(indexRel)

	target, ok := safeJoinSiteRoot(root, ctx.Request.URL.Path)
	if !ok {
		ctx.AbortWithStatus(http.StatusForbidden)
		return
	}

	if fi, err := os.Stat(target); err == nil {
		if !fi.IsDir() {
			ctx.File(target)
			return
		}
		// 请求落在站点根目录：始终使用配置的入口（如 build/index.html）
		if filepath.Clean(target) == root {
			ctx.File(indexPath)
			return
		}
		// 子目录：尝试 子目录/<入口基名>，如 docs/index.html
		dirIndex := filepath.Join(target, indexBase)
		if fi2, err2 := os.Stat(dirIndex); err2 == nil && !fi2.IsDir() {
			ctx.File(dirIndex)
			return
		}
		if site.SPA {
			ctx.File(indexPath)
			return
		}
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}

	// 文件不存在：带扩展名的路径视为具体资源，缺失则 404；无扩展名则交给 SPA 回退
	if path.Ext(ctx.Request.URL.Path) != "" {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	if site.SPA {
		ctx.File(indexPath)
		return
	}
	ctx.AbortWithStatus(http.StatusNotFound)
}
