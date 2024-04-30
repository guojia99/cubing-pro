package basemodel

import (
	"time"

	"gorm.io/gorm"
)

type DBModel interface {
	dbModel()
}

type Model struct {
	ID        uint           `gorm:"column:id;primaryKey" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (m Model) dbModel() {}

type StringIDModel struct {
	ID        string         `gorm:"column:id;primaryKey;type:varchar(64)" json:"id"`
	CreatedAt time.Time      `json:"createdAt"`
	UpdatedAt time.Time      `json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt"`
}

func (m StringIDModel) dbModel() {}
