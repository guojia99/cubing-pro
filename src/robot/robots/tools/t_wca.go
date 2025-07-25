package tools

import (
	"fmt"
	"log"

	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"github.com/guojia99/cubing-pro/src/internel/wca"
	"github.com/guojia99/cubing-pro/src/robot/types"
	"github.com/patrickmn/go-cache"
)

type TWca struct {
	Cache *cache.Cache
}

func (t *TWca) ID() []string {
	return []string{"WCA", "wca", "Wca"}
}

func (t *TWca) Help() string {
	out := "1. 输入 WCA {WcaID} 可查询选手成绩"
	return out
}

var wcaEventsList = []string{
	"333",
	"222",
	"444",
	"555",
	"666",
	"777",
	"333bf",
	"333fm",
	"333oh",
	"clock",
	"minx",
	"pyram",
	"skewb",
	"sq1",
	"444bf",
	"555bf",
	"333mbf",
}

var wcaEventsCnMap = map[string]string{
	"333":    "三阶",
	"222":    "二阶",
	"444":    "四阶",
	"555":    "五阶",
	"666":    "六阶",
	"777":    "七阶",
	"333bf":  "三盲",
	"333fm":  "最少步",
	"333oh":  "单手",
	"clock":  "魔表",
	"minx":   "五魔",
	"pyram":  "金字塔",
	"skewb":  "斜转",
	"sq1":    "SQ-1",
	"444bf":  "四盲",
	"555bf":  "五盲",
	"333mbf": "多盲",
}

func (t *TWca) Do(message types.InMessage) (*types.OutMessage, error) {
	msg := types.RemoveID(message.Message, t.ID())
	msg = utils2.ReplaceAll(msg, "", " ")

	pes, err := wca.ApiSearchPersons(msg)
	if err != nil {
		log.Printf("wca api search persons err: %v", err)
		return message.NewOutMessage("选手查询错误"), nil
	}

	if pes.Total == 0 || len(pes.Rows) == 0 {
		return message.NewOutMessage("查找不到该选手"), nil
	}
	if pes.Total >= 2 {
		out := fmt.Sprintf("查询到%d位选手, 请输入WcaID查询: \n", pes.Total)
		for _, row := range pes.Rows {
			out += fmt.Sprintf("%s %s\n", row.WcaId, row.Name)
		}
		return message.NewOutMessage(out), nil
	}

	personWCAID := pes.Rows[0].WcaId

	result, err := wca.ApiGetWCAResults(personWCAID)
	if err != nil {
		return message.NewOutMessagef("选手%s成绩查询错误", personWCAID), nil
	}
	out := result.PersonName + "\n"
	for _, ev := range wcaEventsList {
		b, hasB := result.Best[ev]
		if !hasB {
			continue
		}
		a, hasA := result.Avg[ev]
		if hasA {
			out += fmt.Sprintf("%s %s | %s\n", wcaEventsCnMap[ev], b.BestStr, a.AverageStr)
		} else {
			out += fmt.Sprintf("%s %s\n", wcaEventsCnMap[ev], b.BestStr)
		}
	}
	return message.NewOutMessage(out), nil
}
