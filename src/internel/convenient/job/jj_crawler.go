package job

import (
	"fmt"
	"log"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/email"
	"github.com/guojia99/cubing-pro/src/robot/crawler"
	"gorm.io/gorm"
)

type JJCrawler struct {
	DB     *gorm.DB
	Config configs.Config
}

func (c *JJCrawler) Name() string {
	return "JJCrawler"
}

func (c *JJCrawler) Run() error {
	find := crawler.CheckAllCubingCompetition()

	log.Printf("cubing获取开始")
	if len(find) > 0 {
		msg := "获取到链接为:\n"
		for _, v := range find {
			msg += fmt.Sprintf("%s\n", v)
		}
		for _, e := range []string{
			"guojia99@foxmail.com",
			"921403690@qq.com",
		} {
			err := email.SendEmail(c.Config.GlobalConfig.EmailConfig, "粗饼爬虫报告", []string{e}, []byte(msg))
			if err != nil {
				log.Printf("[e] 粗饼报告错误 %s\n", err)
			}
		}
	}
	log.Printf("cubing获取结束")
	return nil
}
