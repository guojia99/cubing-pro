package wca

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	wca2 "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/models"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/utils"
	utils2 "github.com/guojia99/cubing-pro/src/internel/utils"
	"gorm.io/gorm"
)

const wcaResultUrlFormat = "https://www.worldcubeassociation.org/api/v0/persons/%s/results" // 2017XUYO01

func ApiGetWCAResults(wcaID string) (*models.PersonBestResults, error) {
	wcaID = strings.ToUpper(wcaID)
	var resp []models.WCAResults
	if err := utils2.HTTPRequestWithJSON(http.MethodGet, fmt.Sprintf(wcaResultUrlFormat, wcaID), nil, nil, nil, &resp); err != nil {
		return nil, err
	}

	var out = models.PersonBestResults{
		PersonName: "",
		Best:       make(map[string]models.Results),
		Avg:        make(map[string]models.Results),
	}

	for _, v := range resp {
		out.PersonName = v.Name
		out.WCAID = v.WcaId

		// 无数据的时候
		if _, ok := out.Best[v.EventId]; (!ok && v.Best > 0) || (v.Best > 0 && ok && utils.IsBestResult(v.EventId, v.Best, out.Best[v.EventId].Best)) {
			out.Best[v.EventId] = models.Results{
				EventId:    v.EventId,
				Best:       v.Best,
				BestStr:    utils.ResultsTimeFormat(v.Best, v.EventId),
				PersonName: v.Name,
				PersonId:   v.WcaId,
			}
		}
		if _, ok := out.Avg[v.EventId]; (!ok && v.Average > 0) || (v.Average > 0 && ok && utils.IsBestResult(v.EventId, v.Average, out.Avg[v.EventId].Average)) {
			out.Avg[v.EventId] = models.Results{
				EventId:    v.EventId,
				Average:    v.Average,
				AverageStr: utils.ResultsTimeFormat(v.Average, v.EventId),
				PersonName: v.Name,
				PersonId:   v.WcaId,
			}
		}

	}

	return &out, nil
}

func GetWcaResultWithDbAndAPI(db *gorm.DB, wcaId string) (*models.PersonBestResults, error) {

	if db == nil {
		return ApiGetWCAResults(wcaId)
	}

	// 从db中查询
	var dbResult wca2.WCAResult
	if err := db.Where("wca_id = ?", wcaId).First(&dbResult).Error; err == nil {
		if time.Since(dbResult.UpdatedAt) <= time.Hour {
			return &dbResult.PersonBestResults, nil
		}
	}

	// api真实查询
	time.Sleep(time.Second)
	res, err := ApiGetWCAResults(wcaId)
	if err != nil {
		return nil, err
	}

	// 缓存到数据库
	if dbResult.ID != 0 {
		dbResult.PersonBestResults = *res
		_ = db.Save(&dbResult)
		return res, nil
	}
	dbResult = wca2.WCAResult{
		WcaID:             wcaId,
		PersonBestResults: *res,
	}
	_ = db.Create(&dbResult)
	return res, nil
}
