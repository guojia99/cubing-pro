package database

import "gorm.io/gorm"

func NewConvenient(db *gorm.DB) ConvenientI {
	return &convenient{
		db: db,
	}
}

type convenient struct {
	db *gorm.DB
}

type ConvenientI interface {
	competitionI
	userI
}
