package gateway

import (
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"regexp"
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

// percentEncodedPathSegments 将各 path 段做 PathEscape，兼容浏览器解码中文 URL 与磁盘上 %XX 目录名。
func percentEncodedPathSegments(urlPath string) string {
	decoded, err := url.PathUnescape(strings.TrimSpace(urlPath))
	if err != nil {
		decoded = urlPath
	}
	trimmed := strings.Trim(decoded, "/")
	if trimmed == "" {
		return "/"
	}
	parts := strings.Split(trimmed, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return "/" + strings.Join(parts, "/")
}

func pathLookupCandidates(urlPath string) []string {
	primary := urlPath
	if primary == "" {
		primary = "/"
	}
	encoded := percentEncodedPathSegments(primary)
	if encoded == primary {
		return []string{primary}
	}
	return []string{primary, encoded}
}

// tryServeResolvedPath 将 URL 映射为文件或子目录 index.html 并响应；成功返回 true。
func tryServeResolvedPath(
	ctx *gin.Context,
	root, indexPath, indexBase, urlPath string,
) bool {
	target, ok := safeJoinSiteRoot(root, urlPath)
	if !ok {
		return false
	}

	fi, err := os.Stat(target)
	if err != nil {
		return false
	}
	if !fi.IsDir() {
		serveFileWithUTF8(ctx, target)
		return true
	}
	if filepath.Clean(target) == filepath.Clean(root) {
		serveFileWithUTF8(ctx, indexPath)
		return true
	}
	dirIndex := filepath.Join(target, indexBase)
	if fi2, err2 := os.Stat(dirIndex); err2 == nil && !fi2.IsDir() {
		serveFileWithUTF8(ctx, dirIndex)
		return true
	}
	return false
}

func tryDynamicRouteFallback(
	ctx *gin.Context,
	root string,
	fallbacks []configs.DynamicRouteFallbackConfig,
) bool {
	if len(fallbacks) == 0 {
		return false
	}

	urlPath := ctx.Request.URL.Path
	for _, fb := range fallbacks {
		match := strings.TrimSpace(fb.Match)
		placeholder := strings.TrimSpace(fb.Placeholder)
		if match == "" || placeholder == "" {
			continue
		}

		re, err := regexp.Compile(match)
		if err != nil || !re.MatchString(urlPath) {
			continue
		}

		target, ok := safeJoinSiteRoot(root, placeholder)
		if !ok {
			continue
		}

		fi, err := os.Stat(target)
		if err != nil || fi.IsDir() {
			continue
		}

		serveFileWithUTF8(ctx, target)
		return true
	}

	return false
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

	for _, candidate := range pathLookupCandidates(ctx.Request.URL.Path) {
		if tryServeResolvedPath(ctx, root, indexPath, indexBase, candidate) {
			return
		}
	}

	urlPath := ctx.Request.URL.Path
	if path.Ext(urlPath) != "" {
		ctx.AbortWithStatus(http.StatusNotFound)
		return
	}
	if site.SPA {
		serveFileWithUTF8(ctx, indexPath)
		return
	}
	if tryDynamicRouteFallback(ctx, root, site.DynamicRouteFallbacks) {
		return
	}
	ctx.AbortWithStatus(http.StatusNotFound)
}

func serveFileWithUTF8(ctx *gin.Context, filePath string) {
	if ct := contentTypeWithUTF8(filePath); ct != "" {
		ctx.Header("Content-Type", ct)
	}
	ctx.File(filePath)
}
