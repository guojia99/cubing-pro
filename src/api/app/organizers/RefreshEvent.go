package organizers

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type RefreshEventReq struct {
	EventID     string `json:"EventID"`
	RoundNumber int    `json:"RoundNumber"` // 轮次
	Open        bool   `json:"Open"`        // 开启
}

func RefreshEvent(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		var req RefreshEventReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 1. 一个项目的轮次开启后， 就得看看这个项目后面的轮次是否已经开启，已经开启了的，就无法给他重新开启。
		// 2. 一个项目关闭后， 自动开启后面的轮次，并计算是否有晋级，并记录对应的晋级数据。

		events := comp.EventMap()
		ev, ok := events[req.EventID]
		if !ok {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "不存在该项目")
			return
		}

		schedule, err := ev.CurRunningSchedule(req.RoundNumber, nil)
		if err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		switch {
		case req.Open && schedule.IsRunning:
			// 开着, 无事发生
			exception.ResponseOK(ctx, nil)
			return
		case !req.Open && !schedule.IsRunning:
			// 关着, 无事发生
			exception.ResponseOK(ctx, nil)
			return
		case req.Open && !schedule.IsRunning:
			// 需要打开
			// 判断是否下一轮已经打开,如果是则不允许打开
			last, err := ev.CurRunningSchedule(req.RoundNumber, utils.Ptr(true))
			if err == nil && last.IsRunning {
				exception.ErrResultUpdate.ResponseWithError(ctx, fmt.Errorf("下一轮 `%s` 已经开启, 本轮将无法开启", last.Round))
				return
			}
			schedule.IsRunning = true
			ev.UpdateSchedule(req.RoundNumber, schedule)
		case !req.Open && schedule.IsRunning:
			// 关闭
			// 判断是否有下一轮, 有的话需要给下一轮更新晋级名单
			schedule.IsRunning = false
			ev.UpdateSchedule(req.RoundNumber, schedule)

			last, err := ev.CurRunningSchedule(req.RoundNumber+1, nil)
			if err == nil {
				var results []result.Results
				svc.DB.
					Where("comp_id = ?", comp.ID).
					Where("round_number = ?", req.RoundNumber).
					Where("event_id = ?", req.EventID).
					Find(&results)
				result.SortResult(results)
				last.AdvancedToThisRound = make([]uint, 0)

				// 部分晋级， 同分的也晋级, 只要排名大于晋级限制人数
				for _, r := range results {
					if last.Competitors == 0 || r.Rank < last.Competitors {
						last.AdvancedToThisRound = append(last.AdvancedToThisRound, r.UserID)
					}
				}

				ev.UpdateSchedule(req.RoundNumber+1, last)
			}

		}

		comp.UpdateEvent(ev)

		if err = svc.DB.Save(&comp).Error; err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
