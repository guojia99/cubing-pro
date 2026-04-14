package gateway

import "github.com/guojia99/cubing-pro/src/configs"

// defaultStaticFileExts 基础静态文件扩展名，配置为空时使用
var defaultStaticFileExts = []string{".css", ".js", ".svg", ".webp", ".woff", ".png", ".jpeg", ".jpg", ".ico", ".json", ".md"}

func getStaticFileExts(cfg configs.GatewayConfig) []string {
	if len(cfg.StaticFileExts) > 0 {
		return cfg.StaticFileExts
	}
	return defaultStaticFileExts
}
