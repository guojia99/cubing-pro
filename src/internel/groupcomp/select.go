package groupcomp

import (
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"gorm.io/gorm"
	"time"
)

func SelectCompetition(db *gorm.DB, groupID, explicitID uint, _ time.Time) (competition.Competition, error) {
	q := db.Where("group_id = ? AND group_id > 0 AND genre = ? AND status = ? AND can_pre_result = ?", groupID, competition.OnlineInformal, competition.Running, true)
	if explicitID != 0 {
		q = q.Where("id = ?", explicitID)
	} else {
		q = q.Where("is_done = ?", false)
	}
	var c competition.Competition
	err := q.Order("created_at ASC, id ASC").First(&c).Error
	return c, err
}

// SelectRobotCompetition preserves the old selection for non-group competition types.
func SelectRobotCompetition(db *gorm.DB, groupID, explicitID uint, now time.Time) (competition.Competition, error) {
	var old competition.Competition
	q := db.Where("group_id = ?", groupID)
	if explicitID != 0 {
		if err := q.Where("id = ?", explicitID).First(&old).Error; err != nil {
			return old, err
		}
		if old.IsGroupCompetition() {
			return SelectCompetition(db, groupID, explicitID, now)
		}
		if old.IsDone {
			return competition.Competition{}, gorm.ErrRecordNotFound
		}
		return old, nil
	}
	var groups int64
	if err := q.Model(&competition.Competition{}).Where("genre = ? AND group_id > 0", competition.OnlineInformal).Count(&groups).Error; err != nil {
		return old, err
	}
	if groups > 0 {
		return SelectCompetition(db, groupID, 0, now)
	}
	q = q.Where("is_done = ?", false)
	if err := q.Order("created_at DESC").First(&old).Error; err != nil {
		return old, err
	}
	return old, nil
}
