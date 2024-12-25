package organizers

import (
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type AddCompResultReq struct {
	CompReq

	//RegId uint `uri:"reg_id"`

	Results  []float64      `json:"Results"`
	CubeID   string         `json:"CubeID"`
	RoundNum int            `json:"Round"`
	EventID  string         `json:"EventId"`
	Penalty  result.Penalty `json:"Penalty"`
}

func checkAndAddPlayerResult(ctx *gin.Context, svc *svc.Svc, req AddCompResultReq) (
	res result.Results, err error,
) {
	// 1. 确认比赛是否存在和是否符合比赛要求和项目是否存在
	// 2. 线上赛的情况下， 自动注册, 并添加对应的项目
	// 3. 其他赛的情况下， 不注册

	//org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)

	// 1. 校验轮次信息
	comp := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
	if !comp.IsRunningTime() {
		err = errors.New("不在比赛时间")
		return
	}
	events := comp.EventMap()
	if _, ok := events[req.EventID]; !ok {
		return result.Results{}, fmt.Errorf("比赛项目不存在")
	}
	ev := events[req.EventID]
	schedule, err := ev.CurRunningSchedule(req.RoundNum, nil)
	if err != nil {
		return result.Results{}, err
	}

	// 2. 获取注册信息
	var usr user.User
	if err = svc.DB.First(&usr, "cube_id = ?", req.CubeID).Error; err != nil {
		return
	}
	var reg competition.Registration
	err = svc.DB.First(&reg, "comp_id = ? and user_id = ?", comp.ID, usr.ID).Error

	switch comp.Genre {
	case competition.OnlineInformal:
		if err != nil {
			reg = competition.Registration{
				CompID:           comp.ID,
				CompName:         comp.Name,
				UserID:           usr.ID,
				UserName:         usr.Name,
				Status:           competition.RegisterStatusPass,
				RegistrationTime: time.Now(),
				AcceptationTime:  utils.PtrTime(time.Now()),
				RetireTime:       nil,
			}
		}
		reg.SetEvent(req.EventID)
		svc.DB.Save(&reg)
		err = nil
	default:
		if err != nil {
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
		if !schedule.FirstRound && !slices.Contains(schedule.AdvancedToThisRound, usr.ID) {
			err = errors.New("不在晋级名单中")
			return
		}
	}

	// 检查上一把是否有成绩
	var lastRes result.Results
	if schedule.RoundNum != 1 {
		err = svc.DB.Where("user_id = ?", usr.ID).Where("comp_id = ?", comp.ID).
			Where("event_id = ?", ev.EventID).Where("round_number = ?", schedule.RoundNum-1).First(&lastRes).Error
		if err != nil || lastRes.ID == 0 {
			err = errors.New("上轮无成绩无法录入")
			return
		}
	}

	req.Results = result.UpdateOrgResult(req.Results, ev.EventRoute, schedule.Cutoff, schedule.CutoffNumber, schedule.TimeLimit)
	err = svc.DB.Where("user_id = ?", usr.ID).
		Where("comp_id = ?", comp.ID).
		Where("event_id = ?", ev.EventID).
		Where("round_number = ?", schedule.RoundNum).
		First(&res).Error

	if err != nil || res.ID == 0 {
		res = result.Results{
			CompetitionID:   comp.ID,
			CompetitionName: comp.Name,
			Round:           schedule.Round,
			RoundNumber:     schedule.RoundNum,
			PersonName:      usr.Name,
			UserID:          usr.ID,
			CubeID:          usr.CubeID,
			EventID:         ev.EventID,
			EventName:       ev.EventName,
			EventRoute:      ev.EventRoute,
			Ban:             false,
			Rank:            0,
		}
	}
	res.Result = req.Results
	res.Penalty = req.Penalty
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
