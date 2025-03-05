package job

import (
	"fmt"
	"log"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/crawler/cubing"
	"github.com/guojia99/cubing-pro/src/internel/database/model/crawler"
	"github.com/guojia99/cubing-pro/src/internel/email"
	"gorm.io/gorm"
)

const wcaCompTemp = `<!DOCTYPE html>
<html lang="zh">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>WCA 比赛列表</title>
    <style>
        body {
            font-family: Arial, sans-serif;
            background-color: #f4f4f4;
            padding: 20px;
        }
        .container {
            max-width: 800px;
            margin: 0 auto;
            background: white;
            padding: 20px;
            border-radius: 8px;
            box-shadow: 0 0 10px rgba(0, 0, 0, 0.1);
        }
        h2 {
            text-align: center;
            color: #333;
        }
        .city {
            margin-bottom: 20px;
        }
        .city h3 {
            background-color: #007bff;
            color: white;
            padding: 10px;
            border-radius: 5px;
        }
        .competition {
            border: 1px solid #ddd;
            margin: 10px 0;
            padding: 10px;
            border-radius: 5px;
            background: #f9f9f9;
        }
        .competition a {
            color: #007bff;
            text-decoration: none;
        }
        .competition a:hover {
            text-decoration: underline;
        }
    </style>
</head>
<body>
    <div class="container">
        <h2>获取到新的 WCA 比赛链接</h2>
        {{range .}}
        <div class="city">
            <h3>{{.City}}</h3>
            {{range .Competitions}}
            <div class="competition">
                <p><strong>比赛名称:</strong> {{.Name}} - {{.Id}}</p>
                <p><strong>比赛项目:</strong> {{range .EventIds}} {{.}} {{end}}</p>
                <p><strong>时间:</strong> {{.StartDate}} - {{.EndDate}}</p>
                <p><strong>链接:</strong> <a href="https://www.worldcubeassociation.org/competitions/{{.Id}}" target="_blank">查看比赛详情</a></p>
            </div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>`

// 比赛信息结构体
type Competition struct {
	Name      string
	Id        string
	EventIds  []string
	StartDate string
	EndDate   string
}

// 城市比赛映射
type CityCompetitions struct {
	City         string
	Competitions []Competition
}

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

		var ccpTmp []CityCompetitions
		for city, cps := range canSendEmail {
			cpTmp := CityCompetitions{
				City:         city,
				Competitions: []Competition{},
			}
			for _, cp := range cps {
				cpTmp.Competitions = append(cpTmp.Competitions, Competition{
					Name:      cp.Name,
					Id:        cp.Id,
					EventIds:  cp.EventIds,
					StartDate: cp.StartDate,
					EndDate:   cp.EndDate,
				})
			}
		}

		if len(needSaveSendEmail) == 0 {
			fmt.Println("无比赛")
			continue
		}

		if err := email.SendEmailWithTemp(c.Config.GlobalConfig.EmailConfig, "粗饼爬虫报告", []string{em}, wcaCompTemp, ccpTmp); err != nil {
			continue
		}
		if err := c.DB.Save(&needSaveSendEmail).Error; err != nil {
			log.Printf("[E] error %s\n", err)
		}
	}
	return nil
}
