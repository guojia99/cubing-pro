package organizers

import (
	"fmt"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type AddCompResultReq struct {
	CompReq

	RegId uint `json:"RegId"`

	Results []float64      `json:"Results"`
	UserId  uint           `json:"UserId"`
	Round   string         `json:"Round"`
	EventID string         `json:"EventId"`
	Penalty result.Penalty `json:"Penalty"`
}

func AddCompResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AddCompResultReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		// 注册审核
		var reg competition.CompetitionRegistration
		if err := svc.DB.First(&reg, "id = ?", req.RegId); err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if !slices.Contains(reg.EventsList(), req.EventID) {
			exception.ErrResultCreate.ResponseWithError(ctx, fmt.Errorf("该选手未报名该项目%s", req.EventID))
			return
		}
		if reg.Status != competition.RegisterStatusPass {
			exception.ErrResultCreate.ResponseWithError(ctx, "该选手比赛资格未审核")
			return
		}

		// 比赛成绩资格
		comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		if !comp.IsRunningTime() {
			exception.ErrResultCreate.ResponseWithError(ctx, "不在比赛时间")
			return
		}
		ev, ok := comp.EventMap()[req.EventID]
		if !ok {
			exception.ErrResultCreate.ResponseWithError(ctx, "本场比赛未开放该项目")
			return
		}
		schedule, err := ev.CurRunningSchedule(req.Round)
		if err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		if !schedule.FirstRound && !slices.Contains(schedule.AdvancedToThisRound, req.UserId) {
			exception.ErrResultCreate.ResponseWithError(ctx, "你不在晋级名单中")
			return
		}

		// todo 成绩是否存在

		// 录入成绩
		res := result.Results{
			CompetitionID: comp.ID,
			Round:         schedule.Round,
			PersonName:    reg.UserName,
			UserID:        reg.UserID,
			Result:        req.Results,
			Penalty:       req.Penalty,
			EventID:       ev.EventID,
			EventName:     ev.EventName,
			EventRoute:    ev.EventRoute,
		}
		if err = res.Update(); err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}

	}
}
