package groupcomp

import (
	"strings"
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func TestEndReadsClockAfterLock(t *testing.T) {
	db, c, _ := fixture(t)
	current := time.Now()
	locked := false
	if err := db.Callback().Query().After("gorm:query").Register("observe_competition_lock", func(tx *gorm.DB) {
		if tx.Statement.Table == "competitions" {
			if _, ok := tx.Statement.Clauses[clause.Locking{}.Name()]; ok {
				locked = true
				current = current.Add(time.Second)
			}
		}
	}); err != nil {
		t.Fatal(err)
	}
	var read time.Time
	if err := End(db, c.ID, func() time.Time {
		if !locked {
			t.Error("clock read before lock")
		}
		read = current
		return read
	}); err != nil {
		t.Fatal(err)
	}
	var got competition.Competition
	if err := db.First(&got, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.CompJSON.ResultClosedAt == nil || !got.CompJSON.ResultClosedAt.Equal(read) {
		t.Fatal("end metadata used pre-lock time")
	}
}

func TestHistoricalParametersOptionalAndCompatible(t *testing.T) {
	for _, confirmed := range []bool{false, true} {
		t.Run(map[bool]string{false: "ordinary approval", true: "legacy confirmed without reason"}[confirmed], func(t *testing.T) {
			db, c, u := fixture(t)
			c.IsDone = true
			c.CompStartTime, c.CompEndTime = time.Time{}, time.Time{}
			c.CompJSON.Events[0].Schedule[0].IsRunning = false
			c.CompJSON.Events[0].Schedule[0].ActualStartTime = time.Time{}
			if err := db.Save(&c).Error; err != nil {
				t.Fatal(err)
			}
			ps := []result.PreResults{pending(c, u)}
			if err := SavePending(db, ps); err != nil {
				t.Fatal(err)
			}
			u.Auth = user.AuthAdmin
			if err := AdminApprove(db, c.ID, ps[0].ID, result.DetailOk, u, confirmed, ""); err != nil {
				t.Fatal(err)
			}
			var p result.PreResults
			if err := db.First(&p, ps[0].ID).Error; err != nil {
				t.Fatal(err)
			}
			if !p.Finish || p.FinishDetail != result.DetailOk {
				t.Fatal("final approval state lost")
			}
			if confirmed && !strings.Contains(p.Detail, "HistoricalConfirmed=true") {
				t.Fatal("compatibility audit lost")
			}
			if !p.CreatedAt.Equal(ps[0].CreatedAt) || !p.UpdatedAt.Equal(ps[0].UpdatedAt) {
				t.Fatal("approval changed submission times")
			}
		})
	}
}

func TestRobotExplicitEndedDefaultActiveOnly(t *testing.T) {
	db, c, _ := fixture(t)
	now := time.Now()
	c.IsDone = true
	c.CreatedAt = now.Add(-time.Hour)
	c.CompStartTime, c.CompEndTime = now.AddDate(0, 0, -4), now.AddDate(0, 0, -3)
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := SelectRobotCompetition(db, c.GroupID, 0, now); err == nil {
		t.Fatal("default selected historical competition")
	}
	got, err := SelectRobotCompetition(db, c.GroupID, c.ID, now)
	if err != nil || got.ID != c.ID {
		t.Fatalf("explicit ended selection failed: %v", err)
	}
	if _, err := SelectRobotCompetition(db, c.GroupID+1, c.ID, now); err == nil {
		t.Fatal("explicit crossed groups")
	}
	active := c
	active.ID = 2
	active.IsDone = false
	active.CreatedAt = now
	active.CompStartTime, active.CompEndTime = now.AddDate(0, 0, 4), now.AddDate(0, 0, 5)
	if err := db.Create(&active).Error; err != nil {
		t.Fatal(err)
	}
	got, err = SelectRobotCompetition(db, c.GroupID, 0, now)
	if err != nil || got.ID != active.ID {
		t.Fatalf("default must ignore dates and skip ended: %v", err)
	}
	c.Status = competition.Reviewing
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := SelectRobotCompetition(db, c.GroupID, c.ID, now); err == nil {
		t.Fatal("explicit unapproved fell back to active")
	}
	c.Status = competition.Running
	c.CanPreResult = false
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := SelectRobotCompetition(db, c.GroupID, c.ID, now); err == nil {
		t.Fatal("explicit disabled fell back")
	}
}
