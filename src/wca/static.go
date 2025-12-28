package wca

import (
	"github.com/guojia99/cubing-pro/src/wca/types"
)

func (w *wca) GetPersonRankTimer(wcaId string) ([]types.StaticPersonRankWithTimer, error) {
	var out []types.StaticPersonRankWithTimer
	if err := w.db.Where("wca_id = ?", wcaId).Find(&out).Error; err != nil {
		return nil, err
	}
	return out, nil
}
