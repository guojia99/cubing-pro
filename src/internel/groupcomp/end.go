package groupcomp

import (
	"errors"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"gorm.io/gorm"
)

// End closes submission, preserving pending results for historical review.
func End(db *gorm.DB, id uint, clock func() time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, id)
		if err != nil {
			return err
		}
		if c.IsDone {
			return nil
		} // Never invent a cutoff for legacy ended competitions.
		if !c.IsGroupCompetition() {
			c.IsDone = true
			return tx.Save(&c).Error
		}
		if clock == nil {
			return errors.New("结束时钟不能为空")
		}
		// Read the cutoff only after obtaining the lock, including prior committed submissions.
		now := clock()
		if c.Status != competition.Running || now.IsZero() {
			return errors.New("比赛未审批或结束时间无效")
		}
		c.IsDone = true
		if c.CompJSON.ResultClosedAt == nil || now.Before(*c.CompJSON.ResultClosedAt) {
			c.CompJSON.ResultClosedAt = &now
		}
		cutoff := *c.CompJSON.ResultClosedAt
		for i := range c.CompJSON.Events {
			for j := range c.CompJSON.Events[i].Schedule {
				s := &c.CompJSON.Events[i].Schedule[j]
				s.IsRunning = false
				if s.ActualEndTime.IsZero() || s.ActualEndTime.After(cutoff) {
					s.ActualEndTime = cutoff
				}
			}
		}
		return tx.Save(&c).Error
	})
}
