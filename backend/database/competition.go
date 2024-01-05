package database

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/guojia99/cubing-pro/core/model"
)

func (c *Convenient) GetCompetitionByName(ctx context.Context, name string) (model.Competition, error) {
	var comp model.Competition

	err := c.db.WithContext(ctx).
		Model(&model.Competition{}).
		Preload(clause.Associations).
		Where("id = ?", name).
		First(&comp).Error

	return comp, err
}
