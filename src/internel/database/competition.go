package database

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/guojia99/cubing-pro/src/internel/database/model/compertion"
)

type competitionI interface {
	GetCompetitionByName(ctx context.Context, name string) (compertion.Competition, error)
}

func (c *convenient) GetCompetitionByName(ctx context.Context, name string) (compertion.Competition, error) {
	var comp compertion.Competition

	err := c.db.WithContext(ctx).
		Model(&compertion.Competition{}).
		Preload(clause.Associations).
		Where("id = ?", name).
		First(&comp).Error

	return comp, err
}
