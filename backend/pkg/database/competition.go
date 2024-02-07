package database

import (
	"context"

	"gorm.io/gorm/clause"

	"github.com/guojia99/cubing-pro/backend/pkg/model/compertion"
)

func (c *Convenient) GetCompetitionByName(ctx context.Context, name string) (compertion.Competition, error) {
	var comp compertion.Competition

	err := c.db.WithContext(ctx).
		Model(&compertion.Competition{}).
		Preload(clause.Associations).
		Where("id = ?", name).
		First(&comp).Error

	return comp, err
}
