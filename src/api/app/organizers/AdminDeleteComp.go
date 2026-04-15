package organizers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
)

// AdminDeleteComp 管理员删除比赛：先删除该比赛下所有成绩（含预录入）、站点纪录、报名与赞助关联，再删除比赛。
func AdminDeleteComp(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		err := svc.DB.Transaction(func(tx *gorm.DB) error {
			var comp competition.Competition
			if err := tx.First(&comp, "id = ?", req.CompId).Error; err != nil {
				return err
			}

			tx.Where("comp_id = ?", req.CompId).Delete(&result.Results{})
			tx.Where("comp_id = ?", req.CompId).Delete(&result.PreResults{})
			tx.Where("comps_id = ?", req.CompId).Delete(&result.Record{})
			tx.Where("comp_id = ?", req.CompId).Delete(&competition.Registration{})
			tx.Where("comp_id = ?", req.CompId).Delete(&competition.AssCompetitionSponsorsUsers{})
			return tx.Delete(&competition.Competition{}, "id = ?", req.CompId).Error
		})
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				exception.ErrResourceNotFound.ResponseWithError(ctx, err)
				return
			}
			exception.ErrResultDelete.ResponseWithError(ctx, err)
			return
		}

		exception.ResponseOK(ctx, nil)
	}
}
