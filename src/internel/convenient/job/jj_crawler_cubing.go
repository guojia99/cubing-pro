package job

import (
	"fmt"
	"log"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing"
	"github.com/guojia99/cubing-pro/src/internel/email"
	"gorm.io/gorm"
)

type JJCrawlerCubing struct {
	DB     *gorm.DB
	Config configs.Config
}

func (c *JJCrawlerCubing) Name() string {
	return "JJCrawlerCubing"
}

func (c *JJCrawlerCubing) Run() error {
	find := cubing.CheckAllCubingCompetition()

	log.Printf("cubing获取开始")
	if len(find) > 0 {
		msg := "获取到链接为:\n"
		for _, v := range find {
			msg += fmt.Sprintf("%s\n", v)
		}
		for _, e := range sendEmails {
			err := email.SendEmail(c.Config.GlobalConfig.EmailConfig, "粗饼爬虫报告", []string{e}, []byte(msg))
			if err != nil {
				log.Printf("[e] 粗饼报告错误 %s\n", err)
			}
		}
	}
	log.Printf("cubing获取结束")
	return nil
}
