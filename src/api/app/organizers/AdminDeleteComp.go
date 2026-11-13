package organizers

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
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
		req.CompId = ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition).ID

		err := svc.DB.Transaction(func(tx *gorm.DB) error {
			if _, err := groupcomp.Lock(tx, req.CompId); err != nil {
				return err
			}

			for _, op := range []struct {
				model  interface{}
				column string
			}{
				{&result.Results{}, "comp_id"}, {&result.PreResults{}, "comp_id"}, {&result.Record{}, "comps_id"}, {&competition.Registration{}, "comp_id"}, {&competition.AssCompetitionSponsorsUsers{}, "comp_id"},
			} {
				if err := tx.Where(op.column+" = ?", req.CompId).Delete(op.model).Error; err != nil {
					return err
				}
			}
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
