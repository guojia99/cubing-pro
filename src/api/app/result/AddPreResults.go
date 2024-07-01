package result

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/api/middleware"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type AddPreResultsReq struct {
	CompId  uint           `json:"CompId"`
	Round   string         `json:"Round"`
	Results []float64      `json:"Results"`
	Penalty result.Penalty `json:"Penalty"`
	EventID string         `json:"EventID"`
}

func AddPreResults(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AddPreResultsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 1. 查看比赛是否存在,
		//	   查看比赛是否可支持预录入 .
		//	   确认项目是否存在，
		//	   确认项目是否已经开启该轮次的该轮次是否有进入到下一轮的资格。
		// 2. 成绩加载。
		// 3. 添加到数据库。
		user, err := middleware.GetAuthUser(ctx)
		if err != nil {
			return
		}

		// 确认报名
		var reg competition.CompetitionRegistration
		if err = svc.DB.First(&reg, "comp_id = ? and user_id = ?", req.CompId, user.ID).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, "你未报名该场比赛")
			return
		}
		if !slices.Contains(reg.EventsList(), req.EventID) {
			exception.ErrResultCreate.ResponseWithError(ctx, fmt.Errorf("你未报名该比赛的%s项目", req.EventID))
			return
		}
		if reg.Status != competition.RegisterStatusPass {
			exception.ErrResultCreate.ResponseWithError(ctx, "你的比赛资格未被审核")
			return
		}

		// 确认比赛
		var comp competition.Competition
		if err = svc.DB.First(&comp, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if !comp.IsRunningTime() {
			exception.ErrResultCreate.ResponseWithError(ctx, "不在比赛时间内")
			return
		}
		if !comp.CanPreResult {
			exception.ErrResultCreate.ResponseWithError(ctx, "本场比赛不允许预录入成绩")
			return
		}
		ev, ok := comp.EventMap()[req.EventID]
		if !ok {
			exception.ErrResultCreate.ResponseWithError(ctx, "本场比赛未开放该项目")
			return
		}

		// 确认比赛是否晋级的资格
		schedule, err := ev.CurRunningSchedule(req.Round, nil)
		if err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		if !schedule.FirstRound && !slices.Contains(schedule.AdvancedToThisRound, user.ID) {
			exception.ErrResultCreate.ResponseWithError(ctx, "你不在晋级名单中")
			return
		}

		// 查看是否有旧的数据存在，如果有则覆盖
		var pre result.PreResults
		if err = svc.DB.First(
			&pre, "comp_id = ? and round = ? and user_id = ? and event_id = ?",
			comp.ID, req.Round, user.ID, ev.EventID,
		).Error; err == nil && pre.ID != 0 {
			if pre.Finish { // 已经处理的就重新生成一个
				pre = result.PreResults{}
			}
		}

		// 写入数据库
		pre = result.PreResults{
			Results: result.Results{
				Model: basemodel.Model{
					ID: pre.ID,
				},
				CompetitionID: comp.ID,
				Round:         schedule.Round,
				PersonName:    user.Name,
				UserID:        user.ID,
				Result:        req.Results,
				Penalty:       req.Penalty,
				EventID:       ev.EventID,
				EventName:     ev.EventName,
				EventRoute:    ev.EventRoute,
			},
			CompsName: comp.Name,
			RoundName: schedule.Round,
			Recorder:  user.Name,
			Source:    "web-api",
		}
		if err = pre.Update(); err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		if err = svc.DB.Save(&req).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
