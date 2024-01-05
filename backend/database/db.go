package database

import "gorm.io/gorm"

type Convenient struct {
	db *gorm.DB
}

type ConvenientI interface {
}
