package configs

import (
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	Debug           bool        `yaml:"debug"`
	Dev             bool        `yaml:"dev"`
	BaseHost        string      `yaml:"baseHost"`
	ImageTempPath   string      `yaml:"imageTempPath"`
	BaseFontTTf     string      `yaml:"baseFontTTf"`
	DB              DBConfig    `yaml:"db"`
	EmailConfig     EmailConfig `yaml:"emailConfig"`
	Scramble        Scramble    `yaml:"scramble"`
	AlgPath         string      `yaml:"algPath"`
	WcaDB           WcaDB       `yaml:"wcaDB"`
	AlgTrainersPath string      `yaml:"algTrainersPath"`
	WcaAuth2        WcaAuth2    `yaml:"wcaAuth2"`
}

type WcaAuth2 struct {
	AppID        string   `yaml:"appId"`
	AppSecret    string   `yaml:"appSecret"`
	RedirectBase string   `yaml:"redirectBase"` // OAuth 回调根地址，如 https://cubing.pro 或 http://localhost:8000
	FrontendBase string   `yaml:"frontendBase"` // 登录成功后跳回的前端域名，如 https://cubing.pro
	RedirectURLs []string `yaml:"redirectURLs"`
	AuthsPath    []string `yaml:"authsPath"`
	Auths        []string `yaml:"auths"` // scopes: public, dob, email, openid, profile
}

type WcaDB struct {
	//SyncUrl string `yaml:"syncUrl"`
	MysqlUrl string `yaml:"mysqlUrl"`
	DbPath   string `yaml:"dbPath"`
	SyncPath string `yaml:"syncPath"`
}

type Scramble struct {
	Type     string `yaml:"type"` // lang, tnoodle
	EndPoint string `yaml:"endpoint"`

	ScrambleDrawType string `yaml:"scrambleDrawType"` // 2mf8
	ScrambleUrl      string `yaml:"scramble"`
}

//type AlgPath struct {
//	Csp       string `yaml:"csp"`
//	CspImages string `yaml:"cspImages"`
//	Bld       string `yaml:"bld"`
//}

type DBConfig struct {
	Driver string `yaml:"driver"`
	DSN    string `yaml:"dsn"`
}

type APIConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	StaticPath string `yaml:"staticPath"`
}

type GatewayConfig struct {
	PEM        string `yaml:"pem"`
	PrivateKey string `yaml:"privateKey"`
	HttpPort   int    `yaml:"httpPort"`
	HTTPSPort  int    `yaml:"httpsPort"`
	HTTPSHost  string `yaml:"httpsHost"`
	XFile      string `yaml:"xFile"`      // 其他特殊文件
	IndexPath  string `yaml:"indexPath"`  // 前端启动文件
	StaticPath string `yaml:"staticPath"` // 其他静态文件

	// StaticFileExts 走静态文件缓存策略的扩展名，空则用代码默认值
	StaticFileExts []string `yaml:"staticFileExts"`

	TNoodlePort        int `yaml:"tNoodlePort"`        // 主机监听tNoodle
	OutsizeTNoodlePort int `yaml:"outsizeTNoodlePort"` // 暴露的tNoodle port

	BldDBPort int `yaml:"blddbPort"`

	// StaticSites 按 Host 托管独立前端静态目录（多子域名 / 多项目）
	StaticSites []StaticSiteConfig `yaml:"staticSites"`
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
	// CacheControl 为空时用 public, max-age=60
	CacheControl string `yaml:"cacheControl"`

	// AutoUpdate 为 true 时：启动 api 且带 -j 时，定时在 RepoDir 执行 git pull，HEAD 变化则执行 BuildCmd（默认 npm run build）。
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

type QQBotConfig struct {
	Enable bool `yaml:"enable"`

	QQ        uint64 `json:"qq,omitempty" toml:"QQ" yaml:"qq"`
	AppId     uint64 `json:"app_id,omitempty" toml:"AppId" yaml:"app_id"`
	Token     string `json:"token,omitempty" toml:"Token" yaml:"token"`
	AppSecret string `json:"app_secret,omitempty" toml:"AppSecret" yaml:"app_secret"`
	IsSandBox bool   `json:"is_sandbox,omitempty" toml:"IsSandBox" yaml:"is_sandbox"`
	WSSAddr   string `json:"wss_addr,omitempty" toml:"WSSAddr" yaml:"wss_addr"`
	IsOpen    bool   `json:"is_open,omitempty" toml:"IsOpen" yaml:"is_open"`

	Server QQBotConfigServer `json:"server" toml:"server" yaml:"server"`
}

type QQBotConfigServer struct {
	Port     int    `yaml:"port"`
	CertFile string `yaml:"certFile"`
	CertKey  string `yaml:"certKey"`
	IsOpen   bool   `yaml:"isOpen"`
}

type WeChatBotConfig struct {
	Enable bool `yaml:"enable"`
}

type CQHttpBot struct {
	Enable  bool   `yaml:"enable"`
	Prefix  string `yaml:"prefix"`  // 命令头 如 .
	Address string `yaml:"address"` // http 地址
	Post    int    `yaml:"post"`    // 反向HTTP地址, 需要robot开启一个地址
}

type RobotConfig struct {
	PersonValPath string `yaml:"personValPath"`

	CQHttpBot []CQHttpBot `yaml:"CQHttpBot"` // cq http qq机器人

	QQBot     []QQBotConfig     `yaml:"QQBot"`
	WeChatBot []WeChatBotConfig `yaml:"WeChatBot"`
}

type EmailConfig struct {
	SmtpHost string `yaml:"smtpHost"`
	SmtpPort int    `yaml:"smtpPort"`
	From     string `yaml:"from"`
	FromName string `yaml:"fromName"`
	Password string `yaml:"password"`
}

type Log struct {
	Path    string `yaml:"path"`
	MaxSize int    `yaml:"maxSize"`
}

type JobConfig struct {
	UnRunJobs []string `yaml:"unRunJobs"`
}

type Config struct {
	Log          Log           `yaml:"log"`
	GlobalConfig GlobalConfig  `yaml:"global"`
	APIConfig    APIConfig     `yaml:"api"`
	Robot        RobotConfig   `yaml:"robot"`
	Gateway      GatewayConfig `yaml:"gateway"`
	Job          JobConfig     `yaml:"job"`
}

func (c *Config) Load(file string) error {
	configBody, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configBody, &c)
	return err
}
