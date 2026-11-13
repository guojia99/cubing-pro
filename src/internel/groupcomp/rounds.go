package groupcomp

import (
	"errors"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
	"time"
)

func Refresh(db *gorm.DB, id uint, eventID string, round int, open bool, now time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, id)
		if err != nil {
			return err
		}
		if !c.IsGroupCompetition() {
			return legacyRefresh(tx, c, eventID, round, open)
		}
		if c.Status != competition.Running || c.IsDone {
			return errors.New("比赛未审批或已经结束")
		}
		ev, ok := c.EventMap()[eventID]
		if !ok {
			return errors.New("项目不存在")
		}
		s, err := ev.CurRunningSchedule(round, nil)
		if err != nil {
			return err
		}
		if s.IsRunning == open {
			return nil
		}
		if s.Competitors < 0 {
			return errors.New("晋级人数不能为负数")
		}
		if open {
			if s.ActualStartTime.IsZero() {
				s.ActualStartTime = now
			}
			s.ActualEndTime = time.Time{}
		} else {
			var pending int64
			if err := tx.Model(&result.PreResults{}).Where("comp_id = ? AND event_id = ? AND round_number = ? AND finish = ?", id, eventID, round, false).Count(&pending).Error; err != nil {
				return err
			}
			if pending > 0 {
				return errors.New("本轮仍有待审批成绩，请先处理")
			}
			s.ActualEndTime = now
			next, err := ev.CurRunningSchedule(round+1, nil)
			if err == nil {
				if next.Competitors < 0 {
					return errors.New("晋级人数不能为负数")
				}
				var rs []result.Results
				if err := tx.Where("comp_id = ? AND event_id = ? AND round_number = ? AND ban = ?", id, eventID, round, false).Find(&rs).Error; err != nil {
					return err
				}
				next.AdvancedToThisRound = Advancing(rs, next.Competitors)
				next.IsRunning = false
				ev.UpdateSchedule(round+1, next)
			}
		}
		s.IsRunning = open
		ev.UpdateSchedule(round, s)
		c.UpdateEvent(ev)
		return tx.Save(&c).Error
	})
}

// OpenAll changes only round availability and timestamps under one competition lock.
func OpenAll(db *gorm.DB, id uint, now time.Time) error {
	return db.Transaction(func(tx *gorm.DB) error {
		c, err := Lock(tx, id)
		if err != nil {
			return err
		}
		if !c.IsGroupCompetition() {
			return errors.New("一键开放仅支持群赛")
		}
		if c.Status != competition.Running || c.IsDone {
			return errors.New("比赛未审批或已经结束")
		}
		if now.IsZero() {
			return errors.New("开放时间无效")
		}
		changed := false
		for i := range c.CompJSON.Events {
			for j := range c.CompJSON.Events[i].Schedule {
				s := &c.CompJSON.Events[i].Schedule[j]
				if !s.IsRunning || s.ActualStartTime.IsZero() || !s.ActualEndTime.IsZero() {
					changed = true
				}
				s.IsRunning = true
				if s.ActualStartTime.IsZero() {
					s.ActualStartTime = now
				}
				s.ActualEndTime = time.Time{}
			}
		}
		if !changed {
			return nil
		}
		return tx.Save(&c).Error
	})
}

func Advancing(rs []result.Results, limit int) []uint {
	result.SortResult(rs)
	out := make([]uint, 0)
	for _, r := range rs {
		if limit == 0 || r.Rank <= limit {
			out = append(out, r.UserID)
		}
	}
	return out
}
