package result

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"time"
)

type PlayerCompsReq struct {
	CubeId string `uri:"cubeId"`
}

type PlayerComp struct {
	ID            uint              `json:"id"`
	Name          string            `json:"Name"`
	StrId         string            `json:"StrId"`
	City          string            `json:"City"`
	Genre         competition.Genre `json:"Genre"`
	CompStartTime time.Time         `json:"CompStartTime,omitempty"`
}

func PlayerComps(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req PlayerCompsReq
		if err := ctx.BindUri(&req); err != nil {
			exception.ErrRequestBinding.ResponseWithError(ctx, err)
			return
		}

		var usr user.User
		if err := svc.DB.First(&usr, "cube_id = ?", req.CubeId).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		var ids []uint
		if err := svc.DB.Model(&competition.Registration{}).Where("user_id = ?", usr.ID).Select("comp_id").Pluck("comp_id", &ids).Error; err != nil {
			exception.ErrDatabase.ResponseWithError(ctx, err)
			return
		}

		var comps []competition.Competition
		var err error
		comps, err = app_utils.GenerallyList[competition.Competition](
			ctx, svc.DB, comps, app_utils.ListSearchParam{
				Model:   &competition.Competition{},
				MaxSize: 0,
				Query:   "id in ?",
				QueryCons: []interface{}{
					ids,
				},
				Select: []string{
					"str_id", "name", "country", "city", "genre",
					"logo", "comp_start_time", "comp_end_time",
					"event_min", "series", "wca_url",
				},
				NotAutoResp: true,
			},
		)
		if err != nil {
			return
		}
		var out []PlayerComp
		for _, comp := range comps {
			out = append(out, PlayerComp{
				ID:            comp.ID,
				Name:          comp.Name,
				StrId:         comp.StrId,
				City:          comp.City,
				Genre:         comp.Genre,
				CompStartTime: comp.CompStartTime,
			})
		}

		exception.ResponseOK(
			ctx, app_utils.GenerallyListResp{
				Items: out,
				Total: int64(len(out)),
			},
		)
	}
}
