package job

import (
	"fmt"
	"log"

	"github.com/guojia99/cubing-pro/src/configs"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing"
	"github.com/guojia99/cubing-pro/src/internel/database/model/crawler"
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

	log.Printf("cubing获取开始")
	find := cubing.NewDCubingCompetition().GetNewCompetitions()

	for _, em := range sendEmails {
		var canSendEmailCp []Competition
		var needSaveSendEmail []crawler.SendEmail

		for _, fid := range find {
			var curSendEmail crawler.SendEmail
			if err := c.DB.Where("type = 'cubing_comps'").Where("`key` = ?", fid.ID).Where("email = ?", em).First(&curSendEmail).Error; err == nil {
				continue
			}
			needSaveSendEmail = append(needSaveSendEmail, crawler.SendEmail{
				Email: em,
				Type:  "cubing_comps",
				Key:   fid.ID,
			})
			canSendEmailCp = append(canSendEmailCp, Competition{
				Name:      fid.Name,
				Id:        fid.ID,
				EventIds:  []string{fid.Events},
				StartDate: fid.Date,
				EndDate:   fid.Date,
				Url:       fmt.Sprintf("https://cubing.com/competition/%s", fid.ID),
			})
		}

		if len(needSaveSendEmail) == 0 {
			continue
		}
		ccp := []CityCompetitions{
			{
				City:         "中国 - 粗饼",
				Competitions: canSendEmailCp,
			},
		}

		if err := email.SendEmailWithTemp(c.Config.GlobalConfig.EmailConfig, "粗饼爬虫报告", []string{em}, wcaCompTemp, ccp); err != nil {
			log.Printf("[E] 发送邮件失败")
			continue
		}
		if err := c.DB.Create(&needSaveSendEmail).Error; err != nil {
			log.Printf("[E] error %s\n", err)
		}
	}
	log.Printf("cubing获取结束")
	return nil
}
