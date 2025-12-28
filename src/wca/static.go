package wca

import (
	"github.com/guojia99/cubing-pro/src/wca/types"
)

func (w *wca) GetPersonRankTimer(wcaId string) ([]types.StaticPersonRankWithTimer, error) {
	var out []types.StaticPersonRankWithTimer
	w.db.Where("wca_id = ?", wcaId).Find(&out)
	return out, nil
}
