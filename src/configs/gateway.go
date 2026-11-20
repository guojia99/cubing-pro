package configs

import "strings"

// DynamicRouteFallbackConfig 未命中物理文件时，将 URL 回退到预构建的 __dynamic__ 占位页（Next 静态导出动态 ID 路由）。
type DynamicRouteFallbackConfig struct {
	Match       string `yaml:"match"`
	Placeholder string `yaml:"placeholder"`
}

type GatewayConfig struct {
	PEM        string `yaml:"pem"`
	PrivateKey string `yaml:"privateKey"`
	HttpPort   int    `yaml:"httpPort"`
	// Port 兼容 yaml 中的 port（与 httpPort 二选一，httpPort 优先）
	Port       int    `yaml:"port"`
	HTTPSPort  int    `yaml:"httpsPort"`
	HTTPSHost  string `yaml:"httpsHost"`
	XFile      string `yaml:"xFile"`      // 其他特殊文件
	IndexPath  string `yaml:"indexPath"`  // 前端启动文件（遗留 Umi 单页，或 Next 多页时指向 dist/index.html）
	StaticPath string `yaml:"staticPath"` // 遗留 Umi：打包资源目录（如 build/static）；Next 多页静态导出请留空并设 spa: false
	// StaticRoot Next.js output:export 等整站静态目录；设置后按路径提供各子目录 index.html（优先于 indexPath/staticPath）
	StaticRoot string `yaml:"staticRoot"`
	// SPA 为 true 时无物理文件则回退 indexPath（Umi/Vue 单页）；Next 多页静态导出须为 false
	SPA bool `yaml:"spa"`

	// DynamicRouteFallbacks Next 多页静态导出：动态 ID 路由回退到 __dynamic__ 占位 HTML
	DynamicRouteFallbacks []DynamicRouteFallbackConfig `yaml:"dynamicRouteFallbacks"`

	// StaticFileExts 走静态文件缓存策略的扩展名，空则用代码默认值
	StaticFileExts []string `yaml:"staticFileExts"`

	TNoodlePort        int `yaml:"tNoodlePort"`        // 主机监听tNoodle
	OutsizeTNoodlePort int `yaml:"outsizeTNoodlePort"` // 暴露的tNoodle port

	BldDBPort int `yaml:"blddbPort"`

	// StaticSites 按 Host 托管独立前端静态目录（多子域名 / 多项目）
	StaticSites []StaticSiteConfig `yaml:"staticSites"`

	// LocalProxies 按 Host 反向代理到本机端口（如开发中的某服务绑定 abc.cubing.pro）
	LocalProxies []LocalProxyConfig `yaml:"localProxies"`
}

// LocalProxyConfig 将请求 Host 反向代理到 TargetHost:Port（默认 127.0.0.1）。
// 可只填 host，或用 hosts 将多个域名指向同一端口。
type LocalProxyConfig struct {
	Host  string   `yaml:"host"`
	Hosts []string `yaml:"hosts"`
	Port  int      `yaml:"port"`
	// TargetHost 上游服务地址，默认 127.0.0.1（如内网 10.x.x.x）
	TargetHost string `yaml:"targetHost"`
}

// StaticSiteConfig 将请求 Host 映射到本地静态资源根目录。
// 可只填 host，或用 hosts 绑定多个域名到同一站点。
type StaticSiteConfig struct {
	Host  string   `yaml:"host"`
	Hosts []string `yaml:"hosts"`
	Root  string   `yaml:"root"`
	// Index 入口文件名，相对 Root，默认 index.html
	Index string `yaml:"index"`
	// SPA 为 true 时：无对应物理文件则回退到 Index（适合 Vue/React 等前端路由）
	SPA bool `yaml:"spa"`
	// DynamicRouteFallbacks 同 GatewayConfig，按站点覆盖（通常仅主站需要）
	DynamicRouteFallbacks []DynamicRouteFallbackConfig `yaml:"dynamicRouteFallbacks"`
	// CacheControl 为空时用 public, max-age=60
	CacheControl string `yaml:"cacheControl"`

	// AutoUpdate 为 true 时：共享服务初始化且带 -j 时，定时在 RepoDir 执行 git pull，HEAD 变化则执行 BuildCmd（默认 npm run build）。
	// RepoDir 为空时使用 Root（适合 root 即仓库根；若 root 仅为 dist 等构建产物目录，请填写 RepoDir 为仓库根路径）。
	AutoUpdate         bool   `yaml:"autoUpdate"`
	AutoUpdateInterval string `yaml:"autoUpdateInterval"` // 如 5m、1h；空则 5m
	RepoDir            string `yaml:"repoDir"`
	BuildCmd           string `yaml:"buildCmd"` // 空则 npm run build
}

// StableID 用于日志与定时任务名：优先首个 host，否则 root。
func (s StaticSiteConfig) StableID() string {
	for _, h := range s.Hosts {
		h = strings.ToLower(strings.TrimSpace(h))
		if h != "" {
			return h
		}
	}
	h := strings.ToLower(strings.TrimSpace(s.Host))
	if h != "" {
		return h
	}
	return strings.TrimSpace(s.Root)
}
