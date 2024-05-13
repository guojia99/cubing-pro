package database

import (
	"context"
	"fmt"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
)

type competitionI interface {
	SearchCompetition(ctx context.Context, searchValue string, genre competition.Genre, startTime, endTime time.Time) ([]competition.Competition, error)
}

// SearchCompetition 查询id、name 符合要求的查询
func (c *convenient) SearchCompetition(ctx context.Context, searchValue string, genre competition.Genre, startTime, endTime time.Time) ([]competition.Competition, error) {
	var out []competition.Competition

	like := fmt.Sprintf("%%%s%%", searchValue)
	sql := c.db.WithContext(ctx).Model(&competition.Competition{}).Limit(100)
	if !startTime.IsZero() {
		sql = sql.Where("comp_start_time > ?", startTime)
	}
	if !endTime.IsZero() {
		sql = sql.Where("comp_start_time < ?", endTime)
	}

	if searchValue != "" {
		sql = sql.Or("id like ?", like).
			Or("str_id like ?", like).
			Or("name like ?", like)
	}

	if genre != 0 {
		sql = sql.Where("genre = ?", genre)
	}

	err := sql.Find(&out).Error
	return out, err
}
