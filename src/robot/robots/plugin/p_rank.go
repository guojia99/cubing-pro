package plugin

import (
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/guojia99/go-tables/table"
)

type RankPlugin struct {
	Svc *svc.Svc
}

var _ types.Plugin = &RankPlugin{}

func (r *RankPlugin) ID() []string {
	return []string{"rank", "排名"}
}

func (r *RankPlugin) Help() string {
	return `获取记录列表:
1. 排名-{项目}-{前N位}: 排名 333 100
`
}

type rankTable struct {
	Rank   int    `table:"排名"`
	SName  string `table:"单次"`
	Single string `table:""`
	Avg    string `table:""`
	AName  string `table:"平均"`
}

func (r *RankPlugin) Do(message types.InMessage) (*types.OutMessage, error) {
	evs := GetEvents(r.Svc, "")
	msg := RemoveID(message.Message, r.ID())

	ev, _, num, err := GetMessageEvent(evs, msg)
	if err != nil {
		return message.NewOutMessage(err.Error()), nil
	}
	if num < 10 {
		num = 10
	}

	bestAll, avgAll := r.Svc.Cov.SelectBestResultsWithEventSort()

	best, ok1 := bestAll[ev.ID]
	avg, _ := avgAll[ev.ID]

	if !ok1 {
		return message.NewOutMessage("暂无成绩"), nil
	}

	var tbs []rankTable
	for i := 0; i < num && i < len(best); i++ {
		var tb = rankTable{
			Rank: i + 1,
		}

		if i < len(best) {
			tb.SName = best[i].PersonName
			tb.Single = best[i].BestString()
		}

		if i < len(avg) {
			tb.AName = avg[i].PersonName
			tb.Avg = avg[i].BestAvgString()
		}
		tbs = append(tbs, tb)
	}

	tb, _ := table.SimpleTable(tbs, &table.Option{
		ExpendID: false,
		Align:    table.AlignLeft,
		Contour:  table.EmptyContour,
	})

	return message.NewOutMessage(tb.String()), nil
}
