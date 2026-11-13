package groupcomp

import (
	"testing"
	"time"

	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"gorm.io/gorm"
)

func TestSubmitEndContinueWriting(t *testing.T) {
	db, c, u := fixture(t)
	ps := []result.PreResults{pending(c, u)}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	id := ps[0].ID
	cutoff := time.Now()
	if err := End(db, c.ID, func() time.Time { return cutoff }); err != nil {
		t.Fatal(err)
	}
	var ended competition.Competition
	if err := db.First(&ended, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !ended.IsDone || ended.CompJSON.ResultClosedAt == nil || !ended.CompJSON.ResultClosedAt.Equal(cutoff) {
		t.Fatal("end metadata missing")
	}
	ps = []result.PreResults{pending(c, u)}
	ps[0].Result = []float64{2, 3, 4, 5, 6}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	if ps[0].ID != id {
		t.Fatal("pending overwrite duplicated after end")
	}
	if err := Approve(db, c.ID, id, result.DetailOk, u); err != nil {
		t.Fatal(err)
	}
	var p result.PreResults
	if err := db.First(&p, id).Error; err != nil {
		t.Fatal(err)
	}
	if !p.Finish || p.ResultID == nil {
		t.Fatal("approval not saved")
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		locked, err := Lock(tx, c.ID)
		if err != nil {
			return err
		}
		_, err = AddResult(tx, locked, u, "333", 1, []float64{3, 4, 5, 6, 7}, nil, time.Now(), true)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteResult(db, c.ID, *p.ResultID); err != nil {
		t.Fatal(err)
	}
	ps = []result.PreResults{pending(c, u)}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	if err := DeletePending(db, ps[0].ID, u.ID); err != nil {
		t.Fatal(err)
	}
	if err := End(db, c.ID, func() time.Time { return cutoff.Add(time.Hour) }); err != nil {
		t.Fatal(err)
	}
	if err := db.First(&ended, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if !ended.CompJSON.ResultClosedAt.Equal(cutoff) {
		t.Fatal("repeat end changed display time")
	}
}

func TestGroupWritingIgnoresTimeAndRoundState(t *testing.T) {
	for _, name := range []string{"before season", "after season", "done", "closed round", "later round opened", "missing times", "zero cutoff", "combined"} {
		t.Run(name, func(t *testing.T) {
			db, c, u := fixture(t)
			now := time.Now()
			s := &c.CompJSON.Events[0].Schedule[0]
			switch name {
			case "before season":
				c.CompStartTime, c.CompEndTime = now.AddDate(0, 0, 2), now.AddDate(0, 0, 3)
			case "after season":
				c.CompStartTime, c.CompEndTime = now.AddDate(0, 0, -3), now.AddDate(0, 0, -2)
			case "done":
				c.IsDone = true
				cutoff := now.Add(-time.Hour)
				c.CompJSON.ResultClosedAt = &cutoff
			case "closed round":
				s.IsRunning = false
				s.ActualStartTime = now.Add(time.Hour)
				s.ActualEndTime = now.Add(-time.Hour)
			case "later round opened":
				c.CompJSON.Events[0].Schedule[1].IsRunning = true
				c.CompJSON.Events[0].Schedule[1].ActualStartTime = now.Add(-time.Hour)
			case "missing times":
				c.IsDone = true
				s.IsRunning = false
				s.ActualStartTime = time.Time{}
				s.ActualEndTime = time.Time{}
			case "zero cutoff":
				zero := time.Time{}
				c.CompJSON.ResultClosedAt = &zero
			case "combined":
				c.IsDone = true
				c.CompStartTime, c.CompEndTime = now.AddDate(0, 0, 2), now.AddDate(0, 0, 3)
				s.IsRunning = false
				s.ActualStartTime = now.Add(time.Hour)
				s.ActualEndTime = now.Add(-time.Hour)
				c.CompJSON.ResultClosedAt = &s.ActualEndTime
				c.CompJSON.Events[0].Schedule[1].ActualStartTime = now.Add(-time.Hour)
			}
			if err := db.Save(&c).Error; err != nil {
				t.Fatal(err)
			}
			ps := []result.PreResults{pending(c, u)}
			if err := SavePending(db, ps); err != nil {
				t.Fatal(err)
			}
			id := ps[0].ID
			ps = []result.PreResults{pending(c, u)}
			ps[0].Result = []float64{2, 3, 4, 5, 6}
			if err := SavePending(db, ps); err != nil {
				t.Fatal(err)
			}
			if ps[0].ID != id {
				t.Fatal("overwrite failed")
			}
			if err := Approve(db, c.ID, id, result.DetailOk, u); err != nil {
				t.Fatal(err)
			}
			var r result.Results
			if err := db.Transaction(func(tx *gorm.DB) error {
				var err error
				r, err = AddResult(tx, c, u, "333", 1, []float64{3, 4, 5, 6, 7}, nil, now, true)
				return err
			}); err != nil {
				t.Fatal(err)
			}
			if err := DeleteResult(db, c.ID, r.ID); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func TestGroupStillRejectsUnapprovedAndCrossCompetition(t *testing.T) {
	db, c, u := fixture(t)
	ps := []result.PreResults{pending(c, u)}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	c.Status = competition.Reviewing
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := SavePending(db, []result.PreResults{pending(c, u)}); err == nil {
		t.Fatal("unapproved pending accepted")
	}
	if err := Approve(db, c.ID, ps[0].ID, result.DetailOk, u); err == nil {
		t.Fatal("unapproved approval accepted")
	}
	if err := db.Transaction(func(tx *gorm.DB) error {
		_, err := AddResult(tx, c, u, "333", 1, []float64{1, 2, 3, 4, 5}, nil, time.Now(), true)
		return err
	}); err == nil {
		t.Fatal("unapproved result accepted")
	}
	if err := DeletePending(db, ps[0].ID, u.ID); err == nil {
		t.Fatal("unapproved deletion accepted")
	}
	c.Status = competition.Running
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	other := c
	other.ID = 2
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	if err := Approve(db, other.ID, ps[0].ID, result.DetailOk, u); err == nil {
		t.Fatal("cross competition approval accepted")
	}
	var r result.Results
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		r, err = AddResult(tx, c, u, "333", 1, []float64{1, 2, 3, 4, 5}, nil, time.Now(), true)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	if err := DeleteResult(db, other.ID, r.ID); err == nil {
		t.Fatal("cross competition result deletion accepted")
	}
	c.Status = competition.Reviewing
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := DeleteResult(db, c.ID, r.ID); err == nil {
		t.Fatal("unapproved result deletion accepted")
	}
}

func TestGroupTimeRelaxationKeepsAdvancementAndPreviousResult(t *testing.T) {
	db, c, u := fixture(t)
	c.IsDone = true
	c.CompJSON.Events[0].Schedule[0].IsRunning = false
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	second := pending(c, u)
	second.RoundNumber = 2
	second.Round = "第二轮"
	if err := SavePending(db, []result.PreResults{second}); err == nil {
		t.Fatal("unqualified second round accepted")
	}
	c.CompJSON.Events[0].Schedule[1].AdvancedToThisRound = []uint{u.ID}
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err := SavePending(db, []result.PreResults{second}); err == nil {
		t.Fatal("second round without previous result accepted")
	}
	var first result.Results
	if err := db.Transaction(func(tx *gorm.DB) error {
		var err error
		first, err = AddResult(tx, c, u, "333", 1, []float64{1, 2, 3, 4, 5}, nil, time.Now(), true)
		return err
	}); err != nil {
		t.Fatal(err)
	}
	ps := []result.PreResults{second}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	if err := Approve(db, c.ID, ps[0].ID, result.DetailOk, u); err != nil {
		t.Fatal(err)
	}
	if err := DeleteResult(db, c.ID, first.ID); err == nil {
		t.Fatal("deleted previous result with dependent later result")
	}
}

func TestReviewFilterCountAndPagination(t *testing.T) {
	db, c, u := fixture(t)
	for _, detail := range []string{result.DetailOk, result.DetailOk, result.DetailNot, result.DetailTimeout, ""} {
		p := pending(c, u)
		p.Finish = detail != ""
		p.FinishDetail = detail
		if err := p.Update(); err != nil {
			t.Fatal(err)
		}
		if err := db.Create(&p).Error; err != nil {
			t.Fatal(err)
		}
	}
	for _, tt := range []struct {
		name, finish, detail, status string
		want                         int64
		bad                          bool
	}{
		{"passed", "", "", "passed", 2, false}, {"rejected", "", "", "rejected", 1, false}, {"timeout", "all", "timeout", "", 1, false}, {"pending", "", "", "", 1, false}, {"all", "all", "", "", 5, false},
		{"conflict", "false", "", "passed", 0, true}, {"alias conflict", "", "not", "passed", 0, true}, {"injection", "", "ok OR 1=1", "", 0, true}, {"invalid status", "", "", "other", 0, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			q, err := FilterReview(db.Model(&result.PreResults{}).Where("comp_id = ?", c.ID), tt.finish, tt.detail, tt.status)
			if tt.bad {
				if err == nil {
					t.Fatal("invalid filter accepted")
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			var total int64
			if err := q.Count(&total).Error; err != nil {
				t.Fatal(err)
			}
			if total != tt.want {
				t.Fatalf("total=%d want %d", total, tt.want)
			}
			var rows []result.PreResults
			if err := q.Order("id ASC").Limit(1).Find(&rows).Error; err != nil {
				t.Fatal(err)
			}
			if len(rows) != 1 {
				t.Fatal("pagination failed")
			}
		})
	}
}
