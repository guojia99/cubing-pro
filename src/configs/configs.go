package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type GlobalConfig struct {
	Debug         bool        `yaml:"debug"`
	Dev           bool        `yaml:"dev"`
	BaseHost      string      `yaml:"baseHost"`
	ImageTempPath string      `yaml:"imageTempPath"`
	BaseFontTTf   string      `yaml:"baseFontTTf"`
	DB            DBConfig    `yaml:"db"`
	EmailConfig   EmailConfig `yaml:"emailConfig"`
	Scramble      Scramble    `yaml:"scramble"`
	AlgPath       string      `yaml:"algPath"`
	WcaDB         WcaDB       `yaml:"wcaDB"`
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

	TNoodlePort        int `yaml:"tNoodlePort"`        // 主机监听tNoodle
	OutsizeTNoodlePort int `yaml:"outsizeTNoodlePort"` // 暴露的tNoodle port
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
