package wca_model

import (
	"strings"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type WCAResult struct {
	basemodel.Model

	WcaID                   string            `gorm:"column:wca_id"`
	PersonBestResults       PersonBestResults `gorm:"-"`
	PersonBestResultsString string
}

func (w *WCAResult) TableName() string { return "wca_results" }

func (w *WCAResult) BeforeSave(*gorm.DB) error {
	w.WcaID = strings.ToUpper(w.WcaID)
	w.PersonBestResultsString, _ = jsoniter.MarshalToString(w.PersonBestResults)
	return nil
}

func (w *WCAResult) AfterFind(tx *gorm.DB) (err error) {
	_ = jsoniter.UnmarshalFromString(w.PersonBestResultsString, &w.PersonBestResults)
	return nil
}
