package basemodel

import (
	"time"

	"gorm.io/gorm"
)

type DBModel interface {
	dbModel()
}

type Model struct {
	ID        uint           `gorm:"column:id;primaryKey" json:"id,omitempty"`
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

func (m Model) dbModel() {}

type StringIDModel struct {
	ID        string         `gorm:"column:id;primaryKey;type:varchar(64)" json:"id,omitempty"`
	CreatedAt time.Time      `json:"createdAt,omitempty"`
	UpdatedAt time.Time      `json:"updatedAt,omitempty"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"deletedAt,omitempty"`
}

func (m StringIDModel) dbModel() {}
