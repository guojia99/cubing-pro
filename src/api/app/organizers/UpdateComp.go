package organizers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
	"strings"
	"time"
)

type UpdateCompReq struct {
	CompReq
	competition.Competition
}

func UpdateComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateCompReq
		current := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		var bindErr error
		if current.IsGroupCompetition() {
			bindErr = ctx.ShouldBindJSON(&req)
		} else {
			bindErr = app_utils.BindAll(ctx, &req)
		}
		if err := bindErr; err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if err := ctx.ShouldBindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		org := ctx.Value(org_mid.OrgAuthMiddlewareKey).(user.Organizers)
		req.CompId = ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition).ID
		var updated competition.Competition
		err := svc.DB.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			comp, err := groupcomp.Lock(tx, req.CompId)
			if err != nil {
				return err
			}
			if comp.OrganizersID != org.ID {
				return errors.New("比赛不属于该团队")
			}
			if comp.IsGroupCompetition() && (req.CompStartTime.IsZero() || req.CompEndTime.IsZero() || req.CompEndTime.In(competition.Beijing).Format("2006-01-02") < req.CompStartTime.In(competition.Beijing).Format("2006-01-02")) {
				return errors.New("比赛日期无效")
			}

			comp.Illustrate = req.Illustrate
			comp.Location = req.Location
			comp.LocationAddr = req.LocationAddr
			comp.Country = req.Country
			comp.City = req.City
			comp.RuleMD = req.RuleMD
			// Round state is only changed through refresh_event, never by editing JSON.
			if comp.IsGroupCompetition() {
				for i := range req.CompJSON.Events {
					ev := &req.CompJSON.Events[i]
					old := comp.EventMap()[ev.EventID]
					for j := range ev.Schedule {
						s := &ev.Schedule[j]
						s.IsRunning = true
						s.ActualStartTime, s.ActualEndTime, s.AdvancedToThisRound = time.Time{}, time.Time{}, nil
						if previous, err := old.CurRunningSchedule(s.RoundNum, nil); err == nil {
							s.IsRunning, s.ActualStartTime, s.ActualEndTime, s.AdvancedToThisRound = previous.IsRunning, previous.ActualStartTime, previous.ActualEndTime, previous.AdvancedToThisRound
						}
					}
				}
			}
			req.CompJSON.ResultClosedAt = comp.CompJSON.ResultClosedAt
			comp.CompJSON = req.CompJSON
			comp.AutomaticReview = req.AutomaticReview
			comp.WCAUrl = req.WCAUrl

			comp.MinCount = req.MinCount
			comp.Count = req.Count
			comp.FreeParticipate = req.FreeParticipate
			if comp.IsGroupCompetition() {
				comp.CanPreResult = req.CanPreResult
			}
			comp.AutomaticReview = req.AutomaticReview

			comp.CompStartTime = req.CompStartTime
			comp.CompEndTime = req.CompEndTime
			comp.RegistrationRestartTime = req.RegistrationRestartTime
			comp.RegistrationEndTime = req.RegistrationEndTime
			comp.RegistrationCancelDeadlineTime = req.RegistrationCancelDeadlineTime
			comp.RegistrationRestartTime = req.RegistrationRestartTime

			if err := tx.Save(&comp).Error; err != nil {
				return err
			}
			updated = comp
			return nil
		})
		if err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, updated)
	}
}

type ApprovalCompReq struct {
	CompReq
	Ok        *bool  `json:"Ok"`
	RejectMsg string `json:"RejectMsg"`
}

func ApprovalComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req ApprovalCompReq
		if err := ctx.ShouldBindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		var existing competition.Competition
		if err := svc.DB.First(&existing, "id = ?", req.CompId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}
		if existing.IsGroupCompetition() {
			if err := ctx.ShouldBindJSON(&req); err != nil {
				exception.ErrRequestBinding.ResponseWithError(ctx, err)
				return
			}
		}
		if err := ctx.ShouldBindUri(&req); err != nil || req.CompId == 0 || (existing.IsGroupCompetition() && req.Ok == nil) {
			exception.ErrRequestBinding.ResponseWithError(ctx, "必须提供比赛ID和Ok")
			return
		}
		err := svc.DB.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			comp, err := groupcomp.Lock(tx, req.CompId)
			if err != nil {
				return err
			}
			if !comp.IsGroupCompetition() {
				if comp.Status == competition.Reviewing {
					comp.Status = competition.Running
					return tx.Save(&comp).Error
				}
				return nil
			}
			if comp.Status != competition.Reviewing {
				return errors.New("比赛不在待审批状态")
			}
			if *req.Ok {
				comp.Status, comp.RejectMsg = competition.Running, ""
			} else {
				if strings.TrimSpace(req.RejectMsg) == "" {
					return errors.New("拒绝原因不能为空")
				}
				comp.Status, comp.RejectMsg = competition.Reject, req.RejectMsg
			}
			return tx.Save(&comp).Error
		})
		if err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
