package job

import (
	"fmt"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing"
	"github.com/guojia99/cubing-pro/src/internel/database/model/crawler"
	"github.com/guojia99/cubing-pro/src/internel/email"
	"gorm.io/gorm"
)

var sendEmails = []string{
	"guojia99@foxmail.com",
	"921403690@qq.com",
}

type JJCrawlerWca struct {
	DB     *gorm.DB
	Config configs.Config
}

func (c *JJCrawlerWca) Name() string {
	return "JJCrawlerWca"
}

func (c *JJCrawlerWca) Run() error {
	curAll := cubing.GetAllWcaComps()
	for _, em := range sendEmails {

		var canSendEmail = make(map[string][]cubing.WcaCompetition)
		var needSaveSendEmail []crawler.SendEmail

		for city, cps := range curAll {
			for _, cp := range cps {
				var curSendEmail crawler.SendEmail
				if err := c.DB.Where("type = ?", "wca_comps").Where("key = ?", cp.Id).Where("email = ?", em).First(&curSendEmail).Error; err == nil {
					continue
				}
				canSendEmail[city] = append(canSendEmail[city], cp)
				needSaveSendEmail = append(needSaveSendEmail, crawler.SendEmail{
					Email: em,
					Type:  "wca_comps",
					Key:   cp.Id,
				})
			}
		}

		var msg = "获取到新的WCA比赛链接如下:\n"

		for city, cps := range canSendEmail {
			msg += fmt.Sprintf("%s --->\n", city)
			for idx, cp := range cps {
				msg += "--------------\n"
				msg += fmt.Sprintf("[%d]比赛名称: %s - %s\n", idx+1, cp.Name, cp.Id)
				msg += fmt.Sprintf("[%d]比赛项目: %+v\n", idx+1, cp.EventIds)
				msg += fmt.Sprintf("[%d]时间%s - %s\n", idx+1, cp.StartDate, cp.EndDate)
				msg += fmt.Sprintf("[%d]链接 https://www.worldcubeassociation.org/competitions/%s\n", idx+1, cp.Id)
			}
			msg += "\n\n"
		}
		if err := email.SendEmail(c.Config.GlobalConfig.EmailConfig, "粗饼爬虫报告", []string{em}, []byte(msg)); err != nil {
			continue
		}
		c.DB.Save(&needSaveSendEmail)
	}
	return nil
}
