package pktimer

import (
	"fmt"
	"sort"
	"strings"
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
	if err := p.Svc.DB.Where("group_id = ? AND running = ?", msg.GroupIDStr(), true).
		First(&pkTimerResult).Error; err != nil {
		return nil
	}

	// 检查是否超时
	if time.Since(pkTimerResult.LastRunning) > pkTimeoutDuration {
		pkTimerResult.Running = false
		_ = p.Svc.DB.Save(&pkTimerResult)
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
	var builder strings.Builder
	builder.WriteString(fmt.Sprintf("%s开启了新一轮的群PK(把数: %d, 项目: %s)\n当前参与玩家:\n",
		results.StartPerson, results.PkResults.Count, results.PkResults.Event.Cn))

	for _, player := range results.PkResults.Players {
		builder.WriteString(fmt.Sprintf("%d. %s\n", player.Index, player.UserName))
	}
	builder.WriteString("----\n")
	builder.WriteString(fmt.Sprintf("本次发货精度为: %.2f%%, 你们快去发货个\n", results.Eps*100))
	builder.WriteString("输入: '加入' 参与本次群PK\n")
	builder.WriteString("输入: '开始' 启动本次群PK\n")
	builder.WriteString("输入: '踢 {编号}' 踢出某个人\n")
	return builder.String()
}

// getMessageTypePrefix 获取消息类型前缀
func getMessageTypePrefix(typ string) string {
	switch typ {
	case "best":
		return "最佳发货成绩:"
	case "avg":
		return "平均发货成绩:"
	default:
		return "这一把:"
	}
}

// getCurResult 获取当前结果
func getCurResult(player pktimerDB.Player, typ string, curNum int) float64 {
	switch typ {
	case "best":
		return player.Best
	case "avg":
		return player.Average
	default:
		if len(player.Results) == curNum {
			return player.Results[curNum-1]
		}
		return 0
	}
}

func getCurPackerMessage(results *pktimerDB.PkTimerResult, typ string) []string {
	curNum := results.PkResults.CurCount
	var players []cachePlayer

	// 收集玩家成绩
	for _, curRes := range results.PkResults.Players {
		curResult := getCurResult(curRes, typ, curNum)
		if curResult == 0 {
			continue
		}
		players = append(players, cachePlayer{
			UserName:  curRes.UserName,
			CurResult: curResult,
		})
	}

	// 按相似成绩分组
	mp := groupPlayersBySimilarScore(players, results.Eps)
	if len(mp) == 0 {
		return nil
	}

	startStr := getMessageTypePrefix(typ)
	var out []string

	for _, ca := range mp {
		if len(ca) == 1 {
			continue
		}

		var builder strings.Builder
		builder.WriteString(startStr)

		// 构建玩家列表
		for _, pl := range ca {
			builder.WriteString(fmt.Sprintf("%s以%s,", pl.UserName, result.TimeParserF2S(pl.CurResult)))
		}

		// 添加发货消息
		if len(ca) == 2 {
			builder.WriteString("发货个！！！\n")
		} else {
			builder.WriteString("多人发货个！！！\n")
		}

		// 计算精度
		n1, n2 := getMinMaxResult(ca)
		pp := getDiffPercent(n1, n2)

		if n1 == n2 {
			builder.WriteString("完美发货个！！！")
		} else if pp <= 0.005 {
			builder.WriteString(fmt.Sprintf("精度: %.2f%%精准发货个!!!\n", pp*100))
		} else {
			builder.WriteString(fmt.Sprintf("精度: %.2f%%发货个!!!\n", pp*100))
		}
		builder.WriteString("\n")

		out = append(out, builder.String())
	}
	return out
}

// getMinMaxResult 获取最小和最大成绩
func getMinMaxResult(ca []cachePlayer) (float64, float64) {
	if len(ca) == 2 {
		if ca[0].CurResult < ca[1].CurResult {
			return ca[0].CurResult, ca[1].CurResult
		}
		return ca[1].CurResult, ca[0].CurResult
	}

	var nps []float64
	for _, pl := range ca {
		nps = append(nps, pl.CurResult)
	}
	sort.Float64s(nps)
	return nps[0], nps[len(nps)-1]
}

func getAllPackerMessage(results *pktimerDB.PkTimerResult) (out []string) {
	bestOut := getCurPackerMessage(results, "best")
	avgOut := getCurPackerMessage(results, "avg")
	out = append(out, bestOut...)
	out = append(out, avgOut...)
	return
}
