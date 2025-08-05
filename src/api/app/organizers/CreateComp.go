package organizers

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type CreateCompReq struct {
	Name  string `json:"Name"`
	StrId string `json:"StrId"`

	Illustrate         string                      `json:"Illustrate"`
	IllustrateHTML     string                      `json:"IllustrateHTML"`
	Location           string                      `json:"Location"`
	Country            string                      `json:"Country"`
	City               string                      `json:"City"`
	RuleMD             string                      `json:"RuleMD"`
	RuleHTML           string                      `json:"RuleHTML"`
	CompJSON           competition.CompetitionJson `json:"CompJSON"`
	Genre              competition.Genre           `json:"genre"`
	Count              int64                       `json:"Count"`
	CanPreResult       bool                        `json:"CanPreResult"`
	CompStartTime      time.Time                   `json:"CompStartTime"`
	CompEndTime        time.Time                   `json:"CompEndTime"`
	GroupID            uint                        `json:"GroupID"`
	CanStartedAddEvent bool                        `json:"CanStartedAddEvent"`

	Apply bool `json:"Apply"`
}

func CreateComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)

		var req CreateCompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		comps := competition.Competition{
			StrId:              req.StrId,
			Status:             competition.Reviewing,
			Name:               req.Name,
			Illustrate:         req.Illustrate,
			IllustrateHTML:     req.IllustrateHTML,
			Location:           req.Location,
			Country:            req.Country,
			City:               req.City,
			RuleMD:             req.RuleMD,
			RuleHTML:           req.RuleHTML,
			CompJSON:           req.CompJSON,
			Genre:              competition.OnlineInformal,
			Count:              req.Count,
			CanPreResult:       true,
			CanStartedAddEvent: req.CanStartedAddEvent,
			CompStartTime:      req.CompStartTime,
			CompEndTime:        req.CompEndTime,
			OrganizersID:       org.ID,
			GroupID:            req.GroupID,
		}
		if req.Apply {
			comps.Status = competition.Reviewing
		}
		if comps.StrId == "" {
			comps.StrId = utils.RandomString(32)
		}

		// 更新comp JSON 打乱
		for i := 0; i < len(comps.CompJSON.Events); i++ {
			ev := comps.CompJSON.Events[i]

			if !ev.IsComp {
				continue
			}
			var eve event.Event
			if err := svc.DB.Where("id = ?", ev.EventID).First(&eve).Error; err != nil {
				continue
			}
			for j := 0; j < len(ev.Schedule); j++ {
				if ev.Schedule[j].NotScramble {
					continue
				}
				comps.CompJSON.Events[i].Schedule[j].Scrambles = make([][]string, 0)
				for k := 0; k < ev.Schedule[j].ScrambleNums; k++ {
					sc, err := svc.Scramble.ScrambleWithComp(eve)
					if err != nil {
						break
					}
					comps.CompJSON.Events[i].Schedule[j].Scrambles = append(comps.CompJSON.Events[i].Schedule[j].Scrambles, sc)
				}
			}
		}

		if err := svc.DB.Create(&comps).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type ApplyCompReq struct {
	CompId uint `uri:"CompId"`
}

func ApplyComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApplyCompReq

		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		var comp competition.Competition
		if err := svc.DB.First(&comp, "id = ? and orgId = ?", req.CompId, org.ID).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		switch comp.Status {
		case competition.Reject, competition.Temporary:
		default:
			exception.ErrResultUpdate.ResponseWithError(ctx, "该状态无法审批")
			return
		}

		comp.Status = competition.Reviewing
		if err := svc.DB.Save(&comp); err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		// todo 发邮件
		exception.ResponseOK(ctx, nil)
	}
}
