package organizers

import (
	"errors"
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

	RegId uint `uri:"reg_id"`

	Results []float64      `json:"Results"`
	UserId  uint           `json:"UserId"`
	Round   string         `json:"Round"`
	EventID string         `json:"EventId"`
	Penalty result.Penalty `json:"Penalty"`
}

func checkAndAddPlayerResult(ctx *gin.Context, svc *svc.Svc, req AddCompResultReq) (
	res result.Results, err error,
) {
	// 注册审核

	var reg competition.Registration
	if err = svc.DB.First(&reg, "id = ?", req.RegId).Error; err != nil {
		return
	}

	if !slices.Contains(reg.EventsList(), req.EventID) {
		err = fmt.Errorf("该选手未报名该项目%s", req.EventID)
		return
	}
	if reg.Status != competition.RegisterStatusPass {
		err = errors.New("该选手比赛资格未审核")
		return
	}

	// 比赛成绩资格
	comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
	if !comp.IsRunningTime() {
		err = errors.New("不在比赛时间")
		return
	}
	ev, ok := comp.EventMap()[req.EventID]
	if !ok {
		err = errors.New("本场比赛未开放该项目")
		return
	}
	schedule, err := ev.CurRunningSchedule(req.Round, nil)
	if err != nil {
		return
	}
	if !schedule.FirstRound && !slices.Contains(schedule.AdvancedToThisRound, req.UserId) {
		err = errors.New("不在晋级名单中")
		return
	}

	// 录入成绩
	req.Results = result.UpdateOrgResult(req.Results, ev.EventRoute, schedule.Cutoff, schedule.CutoffNumber, schedule.TimeLimit)

	res = result.Results{
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
		return
	}

	if err = svc.DB.Save(&res).Error; err != nil {
		return
	}
	return
}

func AddCompResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req AddCompResultReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		_, err := checkAndAddPlayerResult(ctx, svc, req)
		if err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
