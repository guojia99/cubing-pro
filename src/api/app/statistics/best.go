package statistics

import (
	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	_interface "github.com/guojia99/cubing-pro/src/internel/convenient/interface"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
	"github.com/guojia99/cubing-pro/src/internel/utils"
)

type BestReq struct {
	Page int `form:"page" json:"page" query:"page"`
	Size int `form:"size" json:"size" query:"size"`

	EventId string `uri:"eventId"`
	Avg     bool   `form:"avg" json:"avg" query:"avg"`
}

type BestResp struct {
	Best map[_interface.EventID][]result.Results
	Avg  map[_interface.EventID][]result.Results

	BestResults []result.Results
	AvgResults  []result.Results
}

func Best(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req BestReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}

		best, avg := svc.Cov.SelectBestResultsWithEventSort()
		// 无项目
		if req.EventId == "" {
			resp := BestResp{
				Best: make(map[_interface.EventID][]result.Results),
				Avg:  make(map[_interface.EventID][]result.Results),
			}

			for key, bb := range best {
				resp.Best[key] = make([]result.Results, 0)
				rank := bb[0].Rank
				for _, b := range bb {
					if b.Rank == rank {
						resp.Best[key] = append(resp.Best[key], b)
						continue
					}
					break
				}
			}
			for key, aa := range avg {
				resp.Avg[key] = make([]result.Results, 0)
				rank := aa[0].Rank
				for _, a := range aa {
					if a.Rank == rank {
						resp.Avg[key] = append(resp.Avg[key], a)
						continue
					}
					break
				}
			}
			exception.ResponseOK(ctx, resp)
			return
		}

		// 有项目
		resp := BestResp{
			BestResults: make([]result.Results, 0),
			AvgResults:  make([]result.Results, 0),
		}
		if !req.Avg {
			bb, ok := best[req.EventId]
			if !ok {
				exception.ErrResourceNotFound.ResponseWithError(ctx, "无数据")
				return
			}
			resp.BestResults, _ = utils.Page[result.Results](bb, req.Page, req.Size)
			exception.ResponseOK(ctx, resp)
		}
		aa, ok := avg[req.EventId]
		if !ok {
			exception.ErrResourceNotFound.ResponseWithError(ctx, "无数据")
			return
		}
		resp.AvgResults, _ = utils.Page[result.Results](aa, req.Page, req.Size)
		exception.ResponseOK(ctx, resp)

	}

}
