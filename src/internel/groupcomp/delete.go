package groupcomp

import (
	"errors"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
)

func DeletePending(db *gorm.DB, id, userID uint) error {
	var p result.PreResults
	if err := db.First(&p, "id = ? AND user_id = ?", id, userID).Error; err != nil {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, p.CompetitionID)
		if err != nil {
			return err
		}
		if c.IsGroupCompetition() && c.Status != competition.Running {
			return errors.New("比赛未审批")
		}
		op := tx.Where("id = ? AND user_id = ? AND finish = ?", id, userID, false).Delete(&result.PreResults{})
		if op.Error != nil {
			return op.Error
		}
		if op.RowsAffected != 1 {
			return errors.New("成绩已审批或不存在")
		}
		return nil
	})
}

func DeleteResult(db *gorm.DB, compID, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, compID)
		if err != nil {
			return err
		}
		if (!c.IsGroupCompetition() && !c.IsRunningTime()) || (c.IsGroupCompetition() && c.Status != competition.Running) {
			return errors.New("比赛未审批")
		}
		var r result.Results
		if err = tx.First(&r, "id = ? AND comp_id = ?", id, c.ID).Error; err != nil {
			return err
		}
		if c.IsGroupCompetition() {
			ev, ok := c.EventMap()[r.EventID]
			if !ok {
				return errors.New("成绩项目不属于该比赛")
			}
			if _, err := ev.CurRunningSchedule(r.RoundNumber, nil); err != nil {
				return err
			}
		}
		var n int64
		nextQuery := tx.Model(&result.Results{}).Where("comp_id = ? AND event_id = ? AND user_id = ?", c.ID, r.EventID, r.UserID)
		if c.IsGroupCompetition() {
			nextQuery = nextQuery.Where("round_number > ?", r.RoundNumber)
		} else {
			nextQuery = nextQuery.Where("round_number = ?", r.RoundNumber+1)
		}
		if err = nextQuery.Count(&n).Error; err != nil {
			return err
		}
		if n > 0 {
			return errors.New("后续轮次已有成绩，不能删除")
		}
		return tx.Where("id = ? AND comp_id = ?", id, c.ID).Delete(&result.Results{}).Error
	})
}
