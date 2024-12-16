package svc

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	Debug       bool        `yaml:"debug"`
	BaseHost    string      `yaml:"baseHost"`
	XStaticPath string      `yaml:"xStaticPath"`
	XFilePath   string      `yaml:"xFilePath"`
	DB          DBConfig    `yaml:"db"`
	EmailConfig EmailConfig `yaml:"emailConfig"`
}

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
	HTTPSHost  string `yaml:"HTTPSHost"`
	IndexPath  string `yaml:"indexPath"`  // 前端启动文件
	StaticPath string `yaml:"staticPath"` // 其他静态文件
}

type QQBotConfig struct {
	Group     bool     `yaml:"group"`
	Enable    bool     `yaml:"enable"`
	AppID     int      `yaml:"appID"`
	Token     string   `yaml:"token"`
	GroupList []string `yaml:"groupList"`
}

type WeChatBotConfig struct {
	Enable bool `yaml:"enable"`
}

type CQHttpBot struct {
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
