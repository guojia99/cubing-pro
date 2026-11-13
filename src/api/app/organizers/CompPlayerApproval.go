package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"gorm.io/gorm"
	"strconv"
)

type CompPlayerApprovalReq struct {
	Pass *bool `json:"pass"`
}

func CompPlayerApproval(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req CompPlayerApprovalReq
		c := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		var bindErr error
		if c.IsGroupCompetition() {
			bindErr = ctx.ShouldBindJSON(&req)
		} else {
			bindErr = app_utils.BindAll(ctx, &req)
			if req.Pass == nil {
				v := false
				req.Pass = &v
			}
		}
		if err := bindErr; err != nil || req.Pass == nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, "必须提供pass")
			return
		}
		id, err := strconv.ParseUint(ctx.Param("reg_id"), 10, 32)
		if err != nil || id == 0 {
			exception.ErrRequestBinding.ResponseWithError(ctx, "无效报名ID")
			return
		}
		err = svc.DB.WithContext(ctx.Request.Context()).Transaction(func(tx *gorm.DB) error {
			if _, err := groupcomp.Lock(tx, c.ID); err != nil {
				return err
			}
			var reg competition.Registration
			if err := tx.First(&reg, "id = ? AND comp_id = ?", id, c.ID).Error; err != nil {
				return err
			}
			reg.Status = competition.RegisterStatusNotApply
			if *req.Pass {
				reg.Status = competition.RegisterStatusPass
			}
			return tx.Save(&reg).Error
		})
		if err != nil {
			exception.ErrResultUpdate.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, nil)
	}
}
