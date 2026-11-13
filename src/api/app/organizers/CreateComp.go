package organizers

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
	"gorm.io/gorm"
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

	UseTNoodle bool `json:"UseTNoodle"` // 使用tNoodle打乱
	Apply      bool `json:"Apply"`
}

func CreateComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)

		var req CreateCompReq
		if err := ctx.ShouldBind(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var group competition.CompetitionGroup
		if req.GroupID > 0 && svc.DB.First(&group, "id = ? AND orgId = ?", req.GroupID, org.ID).Error != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, "群组不存在或不属于该主办团队")
			return
		}
		if req.GroupID > 0 && (req.CompStartTime.IsZero() || req.CompEndTime.IsZero() || req.CompEndTime.In(competition.Beijing).Format("2006-01-02") < req.CompStartTime.In(competition.Beijing).Format("2006-01-02")) {
			exception.ErrResultCreate.ResponseWithError(ctx, "比赛日期无效")
			return
		}

		comps := competition.Competition{
			StrId:              req.StrId,
			Status:             competition.Temporary,
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
			CanPreResult:       req.CanPreResult,
			CanStartedAddEvent: req.CanStartedAddEvent,
			CompStartTime:      req.CompStartTime,
			CompEndTime:        req.CompEndTime,
			OrganizersID:       org.ID,
			GroupID:            req.GroupID,
		}
		if req.Apply {
			comps.Status = competition.Reviewing
		}
		if !comps.IsGroupCompetition() {
			comps.Status = competition.Reviewing
			comps.CanPreResult = true
		}
		if comps.StrId == "" {
			comps.StrId = utils.RandomString(32)
		}

		// 更新comp JSON 打乱
		var err error
		if req.UseTNoodle {
			comps.CompJSON, err = svc.Scramble.TNoodleScrambles(comps.CompJSON)
		} else {
			comps.CompJSON, err = svc.Scramble.CubingProScrambles(comps.CompJSON)
		}

		if err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		comps.CompJSON.ResultClosedAt = nil
		if comps.IsGroupCompetition() {
			for i := range comps.CompJSON.Events {
				for j := range comps.CompJSON.Events[i].Schedule {
					s := &comps.CompJSON.Events[i].Schedule[j]
					s.IsRunning = true
					s.ActualStartTime, s.ActualEndTime = time.Time{}, time.Time{}
					s.AdvancedToThisRound = nil
				}
			}
		}

		if err = svc.DB.Create(&comps).Error; err != nil {
			exception.ErrResultCreate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}

type ApplyCompReq struct {
	CompId uint `uri:"compId"`
}

func ApplyComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApplyCompReq

		if err := ctx.ShouldBindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		req.CompId = ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition).ID
		err := svc.DB.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			comp, err := groupcomp.Lock(tx, req.CompId)
			if err != nil {
				return err
			}
			if comp.OrganizersID != org.ID {
				return errors.New("比赛不属于该团队")
			}
			if comp.Status != competition.Reject && comp.Status != competition.Temporary {
				return errors.New("该状态无法申请审批")
			}
			comp.Status = competition.Reviewing
			if comp.IsGroupCompetition() {
				comp.RejectMsg = ""
			}
			return tx.Save(&comp).Error
		})
		if err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		// todo 发邮件
		exception.ResponseOK(ctx, nil)
	}
}
