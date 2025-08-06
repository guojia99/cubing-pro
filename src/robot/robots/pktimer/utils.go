package pktimer

import (
	"fmt"
	"sort"
	"time"

	"github.com/2mf8/Better-Bot-Go/log"
	pktimerDB "github.com/guojia99/cubing-pro/src/internel/database/model/pktimer"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/robot/types"
)

func (p *PkTimer) sendMessage(msg *types.OutMessage) {
	err := p.SendMessage(*msg)
	if err != nil {
		log.Error(err)
	}
}

func (p *PkTimer) getMessageDBPkTimer(msg types.InMessage) *pktimerDB.PkTimerResult {
	var pkTimerResult pktimerDB.PkTimerResult
	if p.Svc.DB.Where("group_id = ?", msg.GroupIDStr()).Where("running = ?", true).First(&pkTimerResult).Error != nil {
		return nil
	}

	if time.Since(pkTimerResult.LastRunning) > time.Minute*20 {
		pkTimerResult.Running = false
		p.Svc.DB.Save(&pkTimerResult)
		return nil
	}
	return &pkTimerResult
}

func (p *PkTimer) getMsgUser(msg types.InMessage) (user.User, error) {
	var usr user.User
	var err error
	if msg.QQ != 0 {
		err = p.Svc.DB.Where("qq = ?", msg.QQ).First(&usr).Error
	} else if msg.QQBot != "" {
		err = p.Svc.DB.Where("qq_uni_id = ?", msg.QQBot).First(&usr).Error
	}
	return usr, err
}

func getIniterMessage(results *pktimerDB.PkTimerResult) string {
	out := fmt.Sprintf("%s开启了新一轮的群PK(把数: %d, 项目: %s)\n输入: “加入”参与本次群PK\n当前参与玩家:\n", results.StartPerson, results.PkResults.Count, results.PkResults.Event.Cn)

	for idx, player := range results.PkResults.Players {
		out += fmt.Sprintf("%d. %s\n", idx+1, player.UserName)
	}
	return out
}

func getCurPackerMessage(results *pktimerDB.PkTimerResult, typ string) (out []string) {
	var curNum = results.PkResults.CurCount

	var players []cachePlayer
	for _, curRes := range results.PkResults.Players {
		var pl = cachePlayer{
			UserName: curRes.UserName,
		}
		if typ == "best" {
			pl.CurResult = curRes.Best
		} else if typ == "avg" {
			pl.CurResult = curRes.Average
		} else {
			if len(curRes.Results) != curNum {
				continue
			}
			pl.CurResult = curRes.Results[curNum-1]
		}
		players = append(players, pl)
	}

	mp := groupPlayersBySimilarScore(players)
	if len(mp) == 0 {
		return
	}

	var startStr = "这一把:"
	if typ == "best" {
		startStr = "最佳发货成绩:"
	} else if typ == "avg" {
		startStr = "平均发货成绩:"
	}

	for _, ca := range mp {
		if len(ca) == 1 {
			continue
		}
		curOut := ""
		curOut += startStr
		for _, pl := range ca {
			curOut += fmt.Sprintf("%s以%s,", pl.UserName, result.TimeParserF2S(pl.CurResult))
		}
		if len(ca) == 2 {
			curOut += "发货了！！！\n"
		} else {
			curOut += "多人发货了！！！\n"
		}

		n1, n2 := 0.0, 0.0
		if len(ca) == 2 {
			n1, n2 = ca[0].CurResult, ca[1].CurResult
		} else {
			var nps []float64
			for _, pl := range ca {
				nps = append(nps, pl.CurResult)
			}
			sort.Slice(nps, func(i, j int) bool {
				return nps[i] < nps[j]
			})
			n1, n2 = nps[0], nps[len(nps)-1]
		}

		pp := getDiffPercent(n1, n2)
		if n1 == n2 {
			curOut += fmt.Sprintf("完美发货！！！")
		} else if pp <= 0.005 {
			curOut += fmt.Sprintf("精度: %.2f%%精准发货!!!\n", pp*100)
		} else {
			curOut += fmt.Sprintf("精度: %.2f%%发货了!!!\n", pp*100)
		}
		curOut += "\n"

		out = append(out, curOut)
	}
	return
}

func getAllPackerMessage(results *pktimerDB.PkTimerResult) (out []string) {
	bestOut := getCurPackerMessage(results, "best")
	avgOut := getCurPackerMessage(results, "avg")
	out = append(out, bestOut...)
	out = append(out, avgOut...)
	return
}
