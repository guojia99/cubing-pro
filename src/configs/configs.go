package configs

import (
	"os"

	"gopkg.in/yaml.v3"
)

type APIConfig struct {
	Host       string `yaml:"host"`
	Port       int    `yaml:"port"`
	StaticPath string `yaml:"staticPath"` // API 图片根目录：avatar/、advertisements/；空值使用工作目录，生产路径由部署配置
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
	var raw map[string]interface{}
	if err = yaml.Unmarshal(configBody, &raw); err != nil {
		return err
	}
	// 兼容旧配置键 apiGateway -> gateway
	if _, has := raw["gateway"]; !has {
		if ag, ok := raw["apiGateway"]; ok {
			raw["gateway"] = ag
		}
	}
	out, err := yaml.Marshal(raw)
	if err != nil {
		return err
	}
	if err = yaml.Unmarshal(out, c); err != nil {
		return err
	}
	if c.Gateway.HttpPort == 0 && c.Gateway.Port != 0 {
		c.Gateway.HttpPort = c.Gateway.Port
	}
	return nil
}
