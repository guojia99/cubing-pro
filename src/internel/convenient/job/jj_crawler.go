package job

import (
	"fmt"
	"log"

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

	log.Printf("cubing获取开始")
	if len(find) >= 0 {
		msg := "获取到链接为:\n"
		for _, v := range find {
			msg += fmt.Sprintf("%s\n", v)
		}
		err := email.SendEmail(c.Config.GlobalConfig.EmailConfig, "爬虫报告", []string{"guojia99@foxmail.com"}, []byte(msg))
		return err
	}
	log.Printf("cubing获取结束")
	return nil
}
