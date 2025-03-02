package job

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/email"
	"github.com/guojia99/cubing-pro/src/robot/crawler"
)

type JJCrawler struct {
	Config configs.Config
}

func (c *JJCrawler) Name() string {
	return "JJCrawler"
}

func (c *JJCrawler) Run() error {
	find := crawler.CheckAllCompetition()

	if len(find) >= 0 {
		msg := "获取到链接为:"
		for _, v := range find {
			msg += fmt.Sprintf("%s\n", v)
		}
		err := email.SendEmail(c.Config.GlobalConfig.EmailConfig, "爬虫报告", []string{"guojia99@foxmail.com"}, []byte(msg))
		return err
	}
	return nil
}
