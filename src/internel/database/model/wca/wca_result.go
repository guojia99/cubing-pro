package wca

import (
	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/wca"
	jsoniter "github.com/json-iterator/go"
	"gorm.io/gorm"
)

type WCAResult struct {
	basemodel.Model

	WcaID                   string                `gorm:"column:wca_id"`
	PersonBestResults       wca.PersonBestResults `gorm:"-"`
	PersonBestResultsString string
}

func (w *WCAResult) TableName() string { return "wca_results" }

func (w *WCAResult) BeforeSave(*gorm.DB) error {
	w.PersonBestResultsString, _ = jsoniter.MarshalToString(w.PersonBestResults)
	return nil
}

func (w *WCAResult) AfterFind(tx *gorm.DB) (err error) {
	_ = jsoniter.UnmarshalFromString(w.PersonBestResultsString, &w.PersonBestResults)
	return nil
}
