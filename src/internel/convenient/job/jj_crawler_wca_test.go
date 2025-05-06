package job

import (
	"testing"

	"github.com/guojia99/cubing-pro/src/internel/configs"
	"github.com/guojia99/cubing-pro/src/internel/email"
)

func Test_testTemps(t *testing.T) {
	data := []CityCompetitions{
		{
			City: "北京",
			Competitions: []Competition{
				{
					Name:               "WCA 北京赛 2025",
					Id:                 "beijing2025",
					EventIds:           []string{"3x3x3", "2x2x2", "4x4x4"},
					StartDate:          "2025-05-01",
					EndDate:            "2025-05-03",
					CompetitionCutoffs: []CompetitionCutoff{},
				},
			},
		},
		{
			City: "上海",
			Competitions: []Competition{
				{
					Name:      "WCA 上海赛 2025",
					Id:        "shanghai2025",
					EventIds:  []string{"3x3x3", "5x5x5"},
					StartDate: "2025-06-10",
					EndDate:   "2025-06-12",
				},
			},
		},
	}

	if err := email.SendEmailWithTemp(configs.EmailConfig{
		SmtpHost: "smtp.qq.com",
		SmtpPort: 587,
		From:     "cubingpro@foxmail.com",
		FromName: "cubingPro",
		Password: "apxdjlwmfkxjdhff",
	}, "粗饼爬虫报告-调试", []string{"guojia99@foxmail.com"}, wcaCompTemp, data); err != nil {
		t.Fatalf("send email fail: %v", err)
	}
}

func Test_testTempsReal(t *testing.T) {
	client := &JJCrawlerWca{
		DB: nil,
		Config: configs.Config{
			GlobalConfig: configs.GlobalConfig{
				EmailConfig: configs.EmailConfig{
					SmtpHost: "smtp.qq.com",
					SmtpPort: 587,
					From:     "cubingpro@foxmail.com",
					FromName: "cubingPro",
					Password: "apxdjlwmfkxjdhff",
				},
			},
		},
		debug: true,
	}

	if err := client.Run(); err != nil {
		t.Fatal(err)
	}
}
