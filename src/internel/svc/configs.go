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

type APIGatewayConfig struct {
	PEM        string `yaml:"pem"`
	PrivateKey string `yaml:"privateKey"`
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	APIPort    int    `yaml:"apiPort"`
	HTTPSPort  int    `yaml:"httpsPort"`
	IndexPath  string `yaml:"indexPath"`
	StaticPath string `yaml:"staticPath"`
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

type RobotConfig struct {
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

type Config struct {
	GlobalConfig     GlobalConfig     `yaml:"global"`
	APIGatewayConfig APIGatewayConfig `yaml:"apiGateway"`
	Robot            RobotConfig      `yaml:"robot"`
}

func (c *Config) Load(file string) error {
	configBody, err := os.ReadFile(file)
	if err != nil {
		return err
	}
	err = yaml.Unmarshal(configBody, &c)
	return err
}
