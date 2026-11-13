package groupcomp

import (
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
)

func TestNonGroupDirectRecordingKeepsHEADRules(t *testing.T) {
	for _, tt := range []struct {
		name  string
		genre competition.Genre
		group uint
	}{
		{"online informal no group", competition.OnlineInformal, 0}, {"online official", competition.OnlineOfficial, 0}, {"online official with group", competition.OnlineOfficial, 9}, {"WCA", competition.WCA, 9},
	} {
		t.Run(tt.name, func(t *testing.T) {
			db, c, u := fixture(t)
			c.Genre, c.GroupID = tt.genre, tt.group
			c.IsDone = true
			c.CompStartTime, c.CompEndTime = time.Time{}, time.Time{}
			c.CompJSON.Events[0].Schedule[0].IsRunning = false
			c.CompJSON.Events[0].Schedule[0].ActualStartTime = time.Time{}
			if err := db.Save(&c).Error; err != nil {
				t.Fatal(err)
			}
			if c.Genre != competition.OnlineInformal {
				reg := competition.Registration{CompID: c.ID, UserID: u.ID, Status: competition.RegisterStatusPass}
				reg.SetEvent("333")
				if err := db.Create(&reg).Error; err != nil {
					t.Fatal(err)
				}
			}
			// Closed round, IsDone, zero dates and extra attempts were all accepted by HEAD.
			var r result.Results
			err := db.Transaction(func(tx *gorm.DB) error {
				var err error
				r, err = AddResult(tx, c, u, "333", 1, []float64{1, 2, 3, 4, 5, 6}, nil, time.Now(), true)
				return err
			})
			if err != nil {
				t.Fatal(err)
			}
			if len(r.Result) != 5 {
				t.Fatal("legacy attempt truncation changed")
			}
			if err := Approve(db, c.ID, 999, result.DetailOk, u); err == nil {
				t.Fatal("nonexistent pre accepted")
			}
		})
	}
}

func TestNonGroupWebPendingKeepsRegistrationAndClosedRoundRules(t *testing.T) {
	db, c, u := fixture(t)
	c.GroupID = 0
	c.IsDone = true
	c.CompStartTime, c.CompEndTime = time.Time{}, time.Time{}
	c.CompJSON.Events[0].Schedule[0].IsRunning = false
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := LegacySavePending(db, c, u, "333", "第一轮", []float64{1, 2, 3, 4, 5}, nil); err == nil {
		t.Fatal("web auto-registered non-group player")
	}
	var n int64
	if err := db.Model(&competition.Registration{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("unexpected registration")
	}
	reg := competition.Registration{CompID: c.ID, UserID: u.ID, Status: competition.RegisterStatusPass}
	reg.SetEvent("333")
	if err := db.Create(&reg).Error; err != nil {
		t.Fatal(err)
	}
	p, err := LegacySavePending(db, c, u, "333", "第一轮", []float64{1, 2, 3, 4, 5}, nil)
	if err != nil {
		t.Fatal(err)
	}
	second, err := LegacySavePending(db, c, u, "333", "第一轮", []float64{2, 3, 4, 5, 6}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if second.ID != p.ID {
		t.Fatal("HEAD round-name overwrite changed")
	}
	if err := Approve(db, c.ID, p.ID, result.DetailTimeout, u); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&p, p.ID).Error; err != nil {
		t.Fatal(err)
	}
	if p.Finish {
		t.Fatal("timeout incorrectly became terminal for non-group")
	}
	if err := Approve(db, c.ID, p.ID, result.DetailOk, u); err != nil {
		t.Fatal(err)
	}
}

func TestNonGroupRoundAndEndKeepHEADRules(t *testing.T) {
	db, c, u := fixture(t)
	c.Genre = competition.OnlineOfficial
	c.GroupID = 10
	c.CompJSON.Events[0].Schedule[1].IsRunning = true
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	ps := []result.PreResults{pending(c, u)}
	if err := db.Create(&ps).Error; err != nil {
		t.Fatal(err)
	}
	// HEAD allowed closing with pending records, and did not force-close the next round.
	if err := Refresh(db, c.ID, "333", 1, false, time.Now()); err != nil {
		t.Fatal(err)
	}
	var got competition.Competition
	if err := db.First(&got, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !got.CompJSON.Events[0].Schedule[1].IsRunning || !got.CompJSON.Events[0].Schedule[0].ActualEndTime.IsZero() {
		t.Fatal("new group round rules leaked")
	}
	if err := End(db, c.ID, func() time.Time { t.Error("non-group end should not use cutoff clock"); return time.Now() }); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&got, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !got.IsDone || got.CompJSON.ResultClosedAt != nil || !got.CompJSON.Events[0].Schedule[1].IsRunning {
		t.Fatal("group end state leaked")
	}
	if _, err := SelectCompetition(db, c.GroupID, c.ID, time.Now()); err == nil {
		t.Fatal("robot selected non-group genre")
	}
}

func TestGroupStillEnforcesNewRules(t *testing.T) {
	db, c, u := fixture(t)
	c.CompJSON.Events[0].Schedule[0].IsRunning = false
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := SavePending(db, []result.PreResults{pending(c, u)}); err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		_, err := AddResult(tx, c, u, "333", 1, []float64{1, 2, 3, 4, 5}, nil, time.Now(), true)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	c.CompJSON.Events[0].Schedule[0].IsRunning = true
	c.CompJSON.Events[0].EventRoute = event.RouteType5RoundsAvgHT
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		_, err := AddResult(tx, c, u, "333", 1, []float64{1, 2, 3, 4, 5, 6}, nil, time.Now(), true)
		return err
	}); err == nil {
		t.Fatal("excess group attempts accepted")
	}
}

func TestNonGroupQualificationAndDeleteKeepHEADRules(t *testing.T) {
	db, c, u := fixture(t)
	c.GroupID = 0
	c.CompJSON.Events[0].Schedule[1].IsRunning = false
	c.CompJSON.Events[0].Schedule[1].NoRestrictions = false
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	first, err := LegacyAddResult(db, c, u, "333", 1, []float64{1, 2, 3, 4, 5}, nil)
	if err != nil {
		t.Fatal(err)
	}
	// HEAD OnlineInformal skips advancement membership, including its non-first rounds.
	if _, err := LegacyAddResult(db, c, u, "333", 2, []float64{1, 2, 3, 4, 5}, nil); err != nil {
		t.Fatal(err)
	}
	if err := DeleteResult(db, c.ID, first.ID); err == nil {
		t.Fatal("deleted result with next round result")
	}
	c.Genre = competition.OnlineOfficial
	c.CompJSON.Events[0].Schedule[1].NoRestrictions = true
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if _, err := LegacyAddResult(db, c, u, "333", 2, []float64{1, 2, 3, 4, 5}, nil); err == nil {
		t.Fatal("HEAD official advancement membership rule changed")
	}
}

func TestNonGroupRobotSelectionKeepsHEADOrdering(t *testing.T) {
	db, c, _ := fixture(t)
	c.Genre = competition.OnlineOfficial
	c.Status = competition.Reviewing
	c.CanPreResult = false
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	got, err := SelectRobotCompetition(db, c.GroupID, c.ID, time.Now())
	if err != nil || got.ID != c.ID {
		t.Fatalf("legacy robot selection changed: %v", err)
	}
	if _, err := SelectCompetition(db, c.GroupID, c.ID, time.Now()); err == nil {
		t.Fatal("group selection accepted official online")
	}
}
