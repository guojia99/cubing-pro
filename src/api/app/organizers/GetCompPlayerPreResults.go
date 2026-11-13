package organizers

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/app/organizers/org_mid"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/groupcomp"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"time"
)

type PreListReq struct {
	Finish       string `form:"finish"`
	FinishDetail string `form:"finishDetail"`
	ReviewStatus string `form:"review_status"`
	Event        string `form:"event"`
	Round        int    `form:"round"`
	Player       string `form:"player"`
	Source       string `form:"source"`
	StartTime    int64  `form:"start_time"`
	EndTime      int64  `form:"end_time"`
	Page         int    `form:"page"`
	Size         int    `form:"size"`
}

func GetCompPlayerPreResult(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		c := ctx.Value(org_mid.CompMiddlewareKey).(competition.Competition)
		if !c.IsGroupCompetition() {
			var out []result.PreResults
			app_utils.GenerallyList(ctx, svc.DB, out, app_utils.ListSearchParam[result.PreResults]{Model: &result.PreResults{}, MaxSize: 100, Query: "comp_id = ? AND finish = ?", QueryCons: []interface{}{c.ID, false}, OrderBy: []string{"id"}})
			return
		}
		var req PreListReq
		if err := ctx.ShouldBindQuery(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		q := svc.DB.WithContext(ctx.Request.Context()).Model(&result.PreResults{}).Where("comp_id = ?", c.ID)
		q, err := groupcomp.FilterReview(q, req.Finish, req.FinishDetail, req.ReviewStatus)
		if err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}
		if req.Event != "" {
			q = q.Where("event_id = ?", req.Event)
		}
		if req.Round > 0 {
			q = q.Where("round_number = ?", req.Round)
		}
		if req.Player != "" {
			q = q.Where("(person_name LIKE ? OR cube_id = ?)", "%"+req.Player+"%", req.Player)
		}
		if req.Source != "" {
			q = q.Where("source = ?", req.Source)
		}
		if req.StartTime > 0 {
			q = q.Where("created_at >= ?", time.Unix(req.StartTime, 0))
		}
		if req.EndTime > 0 {
			q = q.Where("created_at < ?", time.Unix(req.EndTime, 0))
		}
		if req.Page < 1 {
			req.Page = 1
		}
		if req.Size < 1 {
			req.Size = 20
		}
		if req.Size > 100 {
			req.Size = 100
		}
		var total int64
		if err := q.Count(&total).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		out := make([]result.PreResults, 0)
		if err := q.Order("created_at ASC, id ASC").Offset((req.Page - 1) * req.Size).Limit(req.Size).Find(&out).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}
		exception.ResponseOK(ctx, app_utils.GenerallyListResp{Items: out, Total: total, Next: int64(req.Page)*int64(req.Size) < total})
	}
}
