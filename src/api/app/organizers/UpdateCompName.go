package organizers

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
)

type UpdateCompNameReq struct {
	CompId  uint   `uri:"compId"`
	NewName string `json:"newName"`
}

func UpdateCompName(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req UpdateCompNameReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
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
			comp.Name = req.NewName
			if err := tx.Save(&comp).Error; err != nil {
				return err
			}
			if err := tx.Model(&result.Results{}).Where("comp_id = ?", req.CompId).UpdateColumn("comp_name", req.NewName).Error; err != nil {
				return err
			}
			if comp.IsGroupCompetition() {
				if err := tx.Model(&result.PreResults{}).Where("comp_id = ?", req.CompId).UpdateColumns(map[string]interface{}{"comp_name": req.NewName, "comps_name": req.NewName}).Error; err != nil {
					return err
				}
			}
			return tx.Model(&competition.Registration{}).Where("comp_id = ?", req.CompId).UpdateColumn("comp_name", req.NewName).Error
		})
		if err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
