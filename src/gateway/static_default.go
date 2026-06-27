package gateway

import (
	"path"
	"path/filepath"
	"slices"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/configs"
)

// serveDefaultStatic 主站默认静态资源（未命中 staticSites 时）。
// - staticRoot 或 indexPath+spa:false：Next 等多页静态导出，按 URL 提供子目录 index.html
// - indexPath+staticPath：遗留 Umi 单页，扩展名走 staticPath，其余回退 indexPath
func serveDefaultStatic(ctx *gin.Context, gw configs.GatewayConfig) {
	if root := strings.TrimSpace(gw.StaticRoot); root != "" {
		serveStaticSite(ctx, configs.StaticSiteConfig{
			Root:                  root,
			Index:                 "index.html",
			SPA:                   gw.SPA,
			DynamicRouteFallbacks: gw.DynamicRouteFallbacks,
		})
		return
	}

	indexPath := strings.TrimSpace(gw.IndexPath)
	staticPath := strings.TrimSpace(gw.StaticPath)

	// Next.js 等多页：index 在产物根目录，无旧版 staticPath 子目录，且显式关闭 SPA 回退
	if indexPath != "" && staticPath == "" && !gw.SPA {
		serveStaticSite(ctx, configs.StaticSiteConfig{
			Root:                  filepath.Dir(indexPath),
			Index:                 filepath.Base(indexPath),
			SPA:                   false,
			DynamicRouteFallbacks: gw.DynamicRouteFallbacks,
		})
		return
	}

	ctx.Header("Cache-Control", "public, max-age=60")
	ext := path.Ext(ctx.Request.URL.Path)
	if staticPath != "" && slices.Contains(getStaticFileExts(gw), ext) {
		serveFileWithUTF8(ctx, filepath.Join(staticPath, ctx.Request.URL.Path))
		return
	}
	if indexPath != "" {
		serveFileWithUTF8(ctx, indexPath)
		return
	}
	ctx.AbortWithStatus(404)
}
