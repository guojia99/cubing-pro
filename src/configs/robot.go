package configs

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
