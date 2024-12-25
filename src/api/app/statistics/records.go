package statistics

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/guojia99/cubing-pro/src/api/exception"
	app_utils "github.com/guojia99/cubing-pro/src/api/utils"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/svc"
)

type RecordsReq struct {
	GroupId string `json:"GroupId"`
	EventId string `json:"EventId"`
}

type RecordsResp struct {
	Records []result.Record

	Best    map[string][]result.Record
	Average map[string][]result.Record
}

func Records(svc *svc.Svc) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var req RecordsReq
		if err := app_utils.BindAll(ctx, &req); err != nil {
			return
		}
		fmt.Println(req)

		var records []result.Record
		db := svc.DB.Model(&result.Record{}).Order("id DESC")
		r := result.RecordTypeWithCubingPro
		if req.GroupId != "" {
			db = db.Where("group_id = ?", req.GroupId)
			r = result.RecordTypeWithGroup
		}
		db = db.Where("d_type = ?", r)

		if req.EventId != "" {
			db = db.Where("event_id = ?", req.EventId)
		}
		if err := db.Find(&records).Error; err != nil {
			exception.ErrResourceNotFound.ResponseWithError(ctx, err)
			return
		}

		if req.EventId != "" {
			exception.ResponseOK(ctx, RecordsResp{Records: records})
			return
		}

		var best = make(map[string][]result.Record)
		var average = make(map[string][]result.Record)

		for _, re := range records {
			if _, ok := best[re.EventId]; !ok && (re.Best != nil || re.Repeatedly != nil) {
				best[re.EventId] = make([]result.Record, 0)
			}
			if _, ok := average[re.EventId]; !ok && re.Average != nil {
				average[re.EventId] = make([]result.Record, 0)
			}

			if re.Repeatedly != nil {
				if len(best[re.EventId]) == 0 || *best[re.EventId][0].Repeatedly == *re.Repeatedly {
					best[re.EventId] = append(best[re.EventId], re)
				}
				continue
			}

			if re.Best != nil {

				if len(best[re.EventId]) == 0 {
					best[re.EventId] = append(best[re.EventId], re)
					continue
				}

				//if re.EventId == "333fm" {
				//	fmt.Println(&best[re.EventId][0].Best == &re.Best, *best[re.EventId][0].Best)
				//}
				if *best[re.EventId][0].Best == *re.Best {
					best[re.EventId] = append(best[re.EventId], re)
				}

				continue
			}

			if re.Average != nil {
				if len(average[re.EventId]) == 0 || *average[re.EventId][0].Average == *re.Average {
					average[re.EventId] = append(average[re.EventId], re)
				}
				continue
			}
		}

		exception.ResponseOK(ctx, RecordsResp{
			Best: best, Average: average,
		})
	}

}
