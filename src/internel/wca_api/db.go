package wca_api

import (
	"time"

	wca2 "github.com/guojia99/cubing-pro/src/internel/database/model/wca"
	"github.com/guojia99/cubing-pro/src/internel/database/wca_model/models"
	"gorm.io/gorm"
)

const dbVersion = "20250908-1520"

func GetWcaResultWithDbAndAPI(db *gorm.DB, wcaId string) (*models.PersonBestResults, error) {

	if db == nil {
		return GetWCAPersonResult(wcaId)
	}

	// 从db中查询
	var dbResult wca2.WCAResult
	if err := db.Where("wca_id = ?", wcaId).First(&dbResult).Error; err == nil {
		if time.Since(dbResult.UpdatedAt) <= time.Hour && dbResult.PersonBestResults.DBVersion == dbVersion {
			return &dbResult.PersonBestResults, nil
		}
	}

	// api真实查询
	time.Sleep(time.Second)
	res, err := GetWCAPersonResult(wcaId)
	if err != nil {
		return nil, err
	}
	res.DBVersion = dbVersion

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
