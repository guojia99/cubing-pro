package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	Debug         bool        `yaml:"debug"`
	BaseHost      string      `yaml:"baseHost"`
	ImageTempPath string      `yaml:"imageTempPath"`
	BaseFontTTf   string      `yaml:"baseFontTTf"`
	DB            DBConfig    `yaml:"db"`
	EmailConfig   EmailConfig `yaml:"emailConfig"`
	Scramble      Scramble    `yaml:"scramble"`
	AlgPath       string      `yaml:"algPath"`
}

type Scramble struct {
	Type     string `yaml:"type"` // lang, tnoodle
	EndPoint string `yaml:"endpoint"`
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
}

type QQBotConfig struct {
	Enable bool `yaml:"enable"`

	QQ        uint64 `json:"qq,omitempty" toml:"QQ" yaml:"qq"`
	AppId     uint64 `json:"app_id,omitempty" toml:"AppId" yaml:"app_id"`
	Token     string `json:"token,omitempty" toml:"Token" yaml:"token"`
	AppSecret string `json:"app_secret,omitempty" toml:"AppSecret" yaml:"app_secret"`
	IsSandBox bool   `json:"is_sandbox,omitempty" toml:"IsSandBox" yaml:"is_sandbox"`
	WSSAddr   string `json:"wss_addr,omitempty" toml:"WSSAddr" yaml:"wss_addr"`
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

type Config struct {
	Log          Log           `yaml:"log"`
	GlobalConfig GlobalConfig  `yaml:"global"`
	APIConfig    APIConfig     `yaml:"api"`
	Robot        RobotConfig   `yaml:"robot"`
	Gateway      GatewayConfig `yaml:"gateway"`
}

func (c *Config) Load(file string) error {
	configBody, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configBody, &c)
	return err
}
