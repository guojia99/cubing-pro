package groupcomp

import (
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestOpenAllPreservesCompetitionData(t *testing.T) {
	for _, partial := range []bool{false, true} {
		t.Run(map[bool]string{false: "all closed", true: "partly open"}[partial], func(t *testing.T) {
			db, c, u := fixture(t)
			now := time.Now().UTC()
			oldStart := now.Add(-time.Hour)
			c.CompStartTime, c.CompEndTime = now.AddDate(0, 0, 3), now.AddDate(0, 0, 4)
			c.CompJSON.Events = append(c.CompJSON.Events, competition.CompetitionEvent{EventID: "444", Schedule: []competition.Schedule{{RoundNum: 1, AdvancedToThisRound: []uint{42}, Scrambles: [][]string{{"R U"}}}}})
			for i := range c.CompJSON.Events {
				for j := range c.CompJSON.Events[i].Schedule {
					s := &c.CompJSON.Events[i].Schedule[j]
					s.IsRunning = false
					s.ActualStartTime = time.Time{}
					s.ActualEndTime = now.Add(-time.Minute)
				}
			}
			if partial {
				s := &c.CompJSON.Events[0].Schedule[0]
				s.IsRunning = true
				s.ActualStartTime = oldStart
				s.ActualEndTime = time.Time{}
			}
			if err := db.Save(&c).Error; err != nil {
				t.Fatal(err)
			}
			p := pending(c, u)
			if err := p.Update(); err != nil {
				t.Fatal(err)
			}
			if err := db.Create(&p).Error; err != nil {
				t.Fatal(err)
			}
			var before competition.Competition
			if err := db.First(&before, c.ID).Error; err != nil {
				t.Fatal(err)
			}
			locks := 0
			if err := db.Callback().Query().After("gorm:query").Register("count_open_all_lock", func(tx *gorm.DB) {
				if tx.Statement.Table == "competitions" {
					if _, ok := tx.Statement.Clauses[clause.Locking{}.Name()]; ok {
						locks++
					}
				}
			}); err != nil {
				t.Fatal(err)
			}
			if err := OpenAll(db, c.ID, now); err != nil {
				t.Fatal(err)
			}
			if locks != 1 {
				t.Fatalf("competition locks=%d", locks)
			}
			var got competition.Competition
			if err := db.First(&got, c.ID).Error; err != nil {
				t.Fatal(err)
			}
			// Build the expected JSON independently to preserve the read-back snapshot.
			var expected competition.Competition
			if err := db.First(&expected, c.ID).Error; err != nil {
				t.Fatal(err)
			}
			for i := range before.CompJSON.Events {
				for j := range before.CompJSON.Events[i].Schedule {
					s := &before.CompJSON.Events[i].Schedule[j]
					s.IsRunning = true
					if s.ActualStartTime.IsZero() {
						s.ActualStartTime = now
					}
					s.ActualEndTime = time.Time{}
				}
			}
			if !reflect.DeepEqual(got.CompJSON, before.CompJSON) {
				t.Fatal("open all modified other JSON fields or incorrect round times")
			}
			var saved result.PreResults
			if err := db.First(&saved, p.ID).Error; err != nil {
				t.Fatal(err)
			}
			if saved.Finish || !reflect.DeepEqual(saved.Result, p.Result) {
				t.Fatal("open all modified pending result")
			}
			updated := got.UpdatedAt
			if err := OpenAll(db, c.ID, now.Add(time.Hour)); err != nil {
				t.Fatal(err)
			}
			var again competition.Competition
			if err := db.First(&again, c.ID).Error; err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(expected.CompJSON, again.CompJSON) || !again.UpdatedAt.Equal(updated) {
				t.Fatal("idempotent all refreshed existing timestamps")
			}
		})
	}
}

func TestOpenAllRejectsWithoutChanges(t *testing.T) {
	for _, name := range []string{"unapproved", "done", "non group", "no group ID", "save failure"} {
		t.Run(name, func(t *testing.T) {
			db, c, _ := fixture(t)
			switch name {
			case "unapproved":
				c.Status = competition.Reviewing
			case "done":
				c.IsDone = true
			case "non group":
				c.Genre = competition.OnlineOfficial
			case "no group ID":
				c.GroupID = 0
			}
			if err := db.Save(&c).Error; err != nil {
				t.Fatal(err)
			}
			var before competition.Competition
			if err := db.First(&before, c.ID).Error; err != nil {
				t.Fatal(err)
			}
			if name == "save failure" {
				if err := db.Callback().Update().Before("gorm:update").Register("fail_open_all", func(tx *gorm.DB) {
					if tx.Statement.Table == "competitions" {
						tx.AddError(errors.New("injected save failure"))
					}
				}); err != nil {
					t.Fatal(err)
				}
			}
			if err := OpenAll(db, c.ID, time.Now()); err == nil {
				t.Fatal("invalid open all succeeded")
			}
			var after competition.Competition
			if err := db.First(&after, c.ID).Error; err != nil {
				t.Fatal(err)
			}
			if before.CompJSONStr != after.CompJSONStr || !before.UpdatedAt.Equal(after.UpdatedAt) {
				t.Fatal("failed open all partially changed competition")
			}
		})
	}
}

func TestSingleGroupOpenIgnoresDateOrderAndOtherOpenRounds(t *testing.T) {
	db, c, _ := fixture(t)
	now := time.Now()
	c.CompStartTime, c.CompEndTime = now.AddDate(0, 0, 2), now.AddDate(0, 0, 3)
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := Refresh(db, c.ID, "333", 2, true, now); err != nil {
		t.Fatal(err)
	}
	var got competition.Competition
	if err := db.First(&got, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	for _, s := range got.CompJSON.Events[0].Schedule {
		if !s.IsRunning {
			t.Fatal("single open closed another round")
		}
	}
	c.IsDone = true
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := Refresh(db, c.ID, "333", 2, true, now); err == nil {
		t.Fatal("single open relaxed ended restriction")
	}
}
