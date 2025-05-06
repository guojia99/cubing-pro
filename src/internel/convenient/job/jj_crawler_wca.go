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

		.cutoff-table {
			width: 100%;
			border-collapse: collapse;
			margin-top: 10px;
		}
		.cutoff-table th, .cutoff-table td {
			border: 1px solid #ccc;
			padding: 8px;
			text-align: center;
		}
		.cutoff-table th {
			background-color: #007bff;
			color: white;
		}
		.cutoff-table tr:nth-child(even) {
			background-color: #f2f2f2;
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
                <p><strong>链接:</strong> <a href="{{.Url}}" target="_blank">查看比赛详情</a></p>

			{{if .CompetitionCutoffs}}
				<p><strong>及格线:</strong></p>
				<table class="cutoff-table">
					<thead>
						<tr>
							<th>项目</th>
							<th>轮次</th>
							<th>及格线(首轮)</th>
							<th>还原时限(首轮)</th>
						</tr>
					</thead>
					<tbody>
						{{range .CompetitionCutoffs}}
						<tr>
							<td>{{.Event}}</td>
							<td>{{.RoundNum}}</td>	
							<td>{{.AttemptResult}}</td>
							<td>{{.LimitResult}}</td>
						</tr>
						{{end}}
					</tbody>
				</table>
			{{end}}

            </div>
            {{end}}
        </div>
        {{end}}
    </div>
</body>
</html>`

type CompetitionCutoff struct {
	Event         string `json:"Event"`
	RoundNum      int    `json:"RoundNum"`      // 轮次
	AttemptResult string `json:"attemptResult"` // 及格线
	LimitResult   string `json:"limitResult"`   // 还原时限
}

// 比赛信息结构体
type Competition struct {
	Name      string
	Id        string
	EventIds  []string
	StartDate string
	EndDate   string
	Url       string

	CompetitionCutoffs []CompetitionCutoff
}

// 城市比赛映射
type CityCompetitions struct {
	City         string
	Competitions []Competition
}

var sendEmails = []string{
	"guojia99@foxmail.com",
	"921403690@qq.com",
	//"yrmfxc@gmail.com",
}

type JJCrawlerWca struct {
	DB     *gorm.DB
	Config configs.Config

	debug bool
}

func (c *JJCrawlerWca) Name() string {
	return "JJCrawlerWca"
}

func AttemptResultString(attemptResult int) string {
	if attemptResult == 0 {
		return "-"
	}

	if attemptResult < 60 {
		return fmt.Sprintf("%d秒", attemptResult)
	}

	hours := attemptResult / 3600
	minutes := (attemptResult % 3600) / 60
	seconds := attemptResult % 60

	if hours > 0 {
		return fmt.Sprintf("%d时%02d分%02d秒", hours, minutes, seconds)
	}

	return fmt.Sprintf("%d分%02d秒", minutes, seconds)
}

func (c *JJCrawlerWca) Run() error {
	curAll := cubing.GetAllWcaComps()
	for _, em := range sendEmails {

		var canSendEmail = make(map[string][]cubing.WcaCompetition)
		var needSaveSendEmail []crawler.SendEmail

		for city, cps := range curAll {
			for _, cp := range cps {

				// 检查邮箱
				if !c.debug {
					var curSendEmail crawler.SendEmail
					err := c.DB.Where("type = 'wca_comps'").Where("`key` = ?", cp.Id).Where("email = ?", em).First(&curSendEmail).Error
					if err == nil || curSendEmail.ID != 0 {
						continue
					}
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
			if len(cps) == 0 {
				continue
			}
			cpTmp := CityCompetitions{
				City:         city,
				Competitions: []Competition{},
			}
			for _, cp := range cps {
				// 获取 WCAInfo
				cpNewTemp := Competition{
					Name:               cp.Name,
					Id:                 cp.Id,
					EventIds:           cp.EventIds,
					StartDate:          cp.StartDate,
					EndDate:            cp.EndDate,
					Url:                fmt.Sprintf("https://www.worldcubeassociation.org/competitions/%s", cp.Id),
					CompetitionCutoffs: make([]CompetitionCutoff, 0),
				}

				info := cubing.GetWcaInfo(cp.Id)
				for _, ev := range info.Events {
					cf := CompetitionCutoff{
						Event:         ev.Id,
						RoundNum:      len(ev.Rounds),
						AttemptResult: "-",
						LimitResult:   AttemptResultString(ev.Rounds[0].TimeLimit.Centiseconds / 100),
					}
					if ev.Rounds[0].Cutoff != nil {
						cf.AttemptResult = AttemptResultString(ev.Rounds[0].Cutoff.AttemptResult / 100)
					}
					cpNewTemp.CompetitionCutoffs = append(cpNewTemp.CompetitionCutoffs, cf)
				}

				cpTmp.Competitions = append(cpTmp.Competitions, cpNewTemp)
			}
			ccpTmp = append(ccpTmp, cpTmp)
		}

		if len(needSaveSendEmail) == 0 {
			fmt.Println("无比赛")
			continue
		}

		subject := "WCA比赛获取报告"
		if c.debug {
			subject = "WCA比赛获取报告-调试"
		}
		if err := email.SendEmailWithTemp(c.Config.GlobalConfig.EmailConfig, subject, []string{em}, wcaCompTemp, ccpTmp); err != nil {
			continue
		}

		if !c.debug {
			if err := c.DB.Create(&needSaveSendEmail).Error; err != nil {
				log.Printf("[E] error %s\n", err)
			}
		}
	}
	return nil
}
