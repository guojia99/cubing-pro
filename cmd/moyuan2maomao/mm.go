package main

import (
	"strings"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// mmp
var (
	dbScr = "root@tcp(127.0.0.1:33036)/cubing_pro?charset=utf8&parseTime=True&loc=Local"
)

const (
	oldKey = "魔缘"
	newKey = "猫猫"

	oldEm = "moyuan@gmail.com"
	newEm = "mm@gmail.com"
)

func main() {
	db, err := gorm.Open(mysql.New(mysql.Config{DSN: dbScr}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		panic(err)
	}

	var comps []competition.Competition
	var groups []competition.CompetitionGroup
	var orgs []user.Organizers

	db.Find(&comps)
	db.Find(&groups)
	db.Find(&orgs)

	for idx, comp := range comps {

		if strings.Contains(comp.Name, oldKey) {
			comps[idx].Name = strings.ReplaceAll(comp.Name, oldKey, newKey)
		}
		if strings.Contains(comp.Illustrate, oldKey) {
			comps[idx].Illustrate = strings.ReplaceAll(comp.Illustrate, oldKey, newKey)
		}
		if strings.Contains(comp.IllustrateHTML, oldKey) {
			comps[idx].IllustrateHTML = strings.ReplaceAll(comp.IllustrateHTML, oldKey, newKey)
		}
		if strings.Contains(comp.RuleMD, oldKey) {
			comps[idx].RuleMD = strings.ReplaceAll(comp.RuleMD, oldKey, newKey)
		}
		if strings.Contains(comp.RuleHTML, oldKey) {
			comps[idx].RuleHTML = strings.ReplaceAll(comp.RuleHTML, oldKey, newKey)
		}
	}
	for idx, group := range groups {

		if strings.Contains(group.Name, oldKey) {
			groups[idx].Name = strings.ReplaceAll(group.Name, oldKey, newKey)
		}
	}
	for idx, org := range orgs {
		if org.Email == oldEm {
			orgs[idx].Email = newEm
		}
		if strings.Contains(org.Name, oldKey) {
			orgs[idx].Name = strings.ReplaceAll(org.Name, oldKey, newKey)
		}
	}

	db.Save(&comps)
	db.Save(&groups)
	db.Save(&orgs)
}
