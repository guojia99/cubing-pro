package basemodel

import (
	"time"

	"gorm.io/gorm"
)

type DBModel interface {
	dbModel()
}

type Model struct {
	ID        uint `gorm:"column:id,primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m Model) dbModel() {}

type StringIDModel struct {
	ID        string `gorm:"column:id,primaryKey"`
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
}

func (m StringIDModel) dbModel() {}
