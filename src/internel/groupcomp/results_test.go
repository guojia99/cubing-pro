package groupcomp

import (
	"errors"
	"reflect"
	"slices"
	"sync"
	"testing"
	"time"

	basemodel "github.com/guojia99/cubing-pro/src/internel/database/model/base"
	"github.com/guojia99/cubing-pro/src/internel/database/model/competition"
	"github.com/guojia99/cubing-pro/src/internel/database/model/event"
	"github.com/guojia99/cubing-pro/src/internel/database/model/result"
	"github.com/guojia99/cubing-pro/src/internel/database/model/user"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// Only an isolated in-memory database is used. No svc/config/init or production DSN.
func fixture(t *testing.T) (*gorm.DB, competition.Competition, user.User) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{Logger: logger.Default.LogMode(logger.Silent)})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatal(err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { sqlDB.Close() })
	if err = db.AutoMigrate(&competition.Competition{}, &competition.Registration{}, &user.User{}, &result.Results{}, &result.PreResults{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now()
	c := competition.Competition{Model: basemodel.Model{ID: 1}, Name: "group test", Status: competition.Running, Genre: competition.OnlineInformal, GroupID: 7, CanPreResult: true, CompStartTime: now.AddDate(0, 0, -1), CompEndTime: now.AddDate(0, 0, 1), CompJSON: competition.CompetitionJson{Events: []competition.CompetitionEvent{{EventID: "333", EventName: "三阶", EventRoute: event.RouteType5RoundsAvgHT, Schedule: []competition.Schedule{{Round: "第一轮", RoundNum: 1, FirstRound: true, IsRunning: true, ActualStartTime: now.Add(-time.Hour)}, {Round: "第二轮", RoundNum: 2, Competitors: 2}}}}}}
	u := user.User{Model: basemodel.Model{ID: 1}, CubeID: "test-player", Name: "Test Player"}
	if err = db.Create(&c).Error; err != nil {
		t.Fatal(err)
	}
	if err = db.Create(&u).Error; err != nil {
		t.Fatal(err)
	}
	return db, c, u
}

func pending(c competition.Competition, u user.User) result.PreResults {
	return result.PreResults{Results: result.Results{CompetitionID: c.ID, CompetitionName: c.Name, Round: "第一轮", RoundNumber: 1, UserID: u.ID, CubeID: u.CubeID, PersonName: u.Name, EventID: "333", EventName: "三阶", EventRoute: event.RouteType5RoundsAvgHT, Result: []float64{1, 2, 3, 4, 5}}, Source: "web-api"}
}

func TestPendingOverwriteApprovalAndScope(t *testing.T) {
	db, c, u := fixture(t)
	other := c
	other.ID = c.ID + 1
	if err := db.Create(&other).Error; err != nil {
		t.Fatal(err)
	}
	pres := []result.PreResults{pending(c, u)}
	if err := SavePending(db, pres); err != nil {
		t.Fatal(err)
	}
	id := pres[0].ID
	pres = []result.PreResults{pending(c, u)}
	pres[0].Result = []float64{2, 3, 4, 5, 6}
	if err := SavePending(db, pres); err != nil {
		t.Fatal(err)
	}
	if pres[0].ID != id {
		t.Fatal("pending was duplicated")
	}
	if err := Approve(db, c.ID+1, id, result.DetailOk, u); err == nil {
		t.Fatal("cross competition approval accepted")
	}
	if err := DeletePending(db, id, u.ID+1); err == nil {
		t.Fatal("cross owner deletion accepted")
	}
	if err := Approve(db, c.ID, id, "invalid", u); err == nil {
		t.Fatal("invalid approval accepted")
	}
	if err := Approve(db, c.ID, id, result.DetailOk, u); err != nil {
		t.Fatal(err)
	}
	var p result.PreResults
	if err := db.First(&p, id).Error; err != nil {
		t.Fatal(err)
	}
	if !p.Finish || p.ResultID == nil || *p.ResultID == 0 {
		t.Fatalf("approval incomplete: %+v", p)
	}
	if err := Approve(db, c.ID, id, result.DetailOk, u); err == nil {
		t.Fatal("duplicate approval accepted")
	}
	if err := DeleteResult(db, c.ID+1, *p.ResultID); err == nil {
		t.Fatal("cross competition result deletion accepted")
	}
	if err := DeletePending(db, id, u.ID); err == nil {
		t.Fatal("approved pending deleted")
	}
	pres = []result.PreResults{pending(c, u)}
	if err := SavePending(db, pres); err != nil {
		t.Fatal(err)
	}
	if pres[0].ID == id {
		t.Fatal("completed history overwritten")
	}
}

func TestBatchRollback(t *testing.T) {
	db, c, u := fixture(t)
	a, b := pending(c, u), pending(c, u)
	b.EventID = "missing"
	if err := SavePending(db, []result.PreResults{a, b}); err == nil {
		t.Fatal("invalid batch accepted")
	}
	var n int64
	db.Model(&result.PreResults{}).Count(&n)
	if n != 0 {
		t.Fatal("partial pending write")
	}
	db.Model(&competition.Registration{}).Count(&n)
	if n != 0 {
		t.Fatal("partial registration write")
	}
	if err := db.Callback().Create().Before("gorm:create").Register("fail_pending", func(tx *gorm.DB) {
		if tx.Statement.Table == "pre_results" {
			tx.AddError(errors.New("injected save error"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := SavePending(db, []result.PreResults{a}); err == nil {
		t.Fatal("save failure ignored")
	}
	db.Model(&competition.Registration{}).Count(&n)
	if n != 0 {
		t.Fatal("registration survived failed save")
	}
}

func TestOldPendingAfterEndDate(t *testing.T) {
	db, c, u := fixture(t)
	yesterday := time.Now().In(competition.Beijing).AddDate(0, 0, -1)
	c.CompStartTime, c.CompEndTime = yesterday, yesterday
	s := &c.CompJSON.Events[0].Schedule[0]
	s.ActualStartTime = yesterday.Add(-time.Hour)
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	p := pending(c, u)
	p.CreatedAt, p.UpdatedAt = yesterday, yesterday
	if err := p.Update(); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&p).Error; err != nil {
		t.Fatal(err)
	}
	if err := SavePending(db, []result.PreResults{pending(c, u)}); err != nil {
		t.Fatalf("submission after date end rejected: %v", err)
	}
	if err := Approve(db, c.ID, p.ID, result.DetailDoubtful, u); err != nil {
		t.Fatal(err)
	}
	if err := Approve(db, c.ID, p.ID, result.DetailOk, u); err != nil {
		t.Fatalf("valid old pending rejected: %v", err)
	}
	bad := pending(c, u)
	bad.RoundNumber = 0
	bad.CreatedAt, bad.UpdatedAt = yesterday, yesterday
	if err := bad.Update(); err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&bad).Error; err != nil {
		t.Fatal(err)
	}
	if err := Approve(db, c.ID, bad.ID, result.DetailOk, u); err == nil {
		t.Fatal("invalid historical pending approved")
	}
}

func TestConcurrentApproval(t *testing.T) {
	db, c, u := fixture(t)
	ps := []result.PreResults{pending(c, u)}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	errs := make(chan error, 2)
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() { defer wg.Done(); errs <- Approve(db, c.ID, ps[0].ID, result.DetailOk, u) }()
	}
	wg.Wait()
	close(errs)
	success := 0
	for err := range errs {
		if err == nil {
			success++
		}
	}
	if success != 1 {
		t.Fatalf("successful approvals = %d", success)
	}
	var n int64
	db.Model(&result.Results{}).Count(&n)
	if n != 1 {
		t.Fatalf("results = %d", n)
	}
}

func TestSelectionOldestAndExplicit(t *testing.T) {
	db, c, _ := fixture(t)
	now := time.Now()
	// Equal creation timestamps must use ascending ID.
	c.CreatedAt = now.Add(-time.Hour)
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	c2 := c
	c2.ID = 2
	if err := db.Create(&c2).Error; err != nil {
		t.Fatal(err)
	}
	got, err := SelectCompetition(db, 7, 0, now)
	if err != nil || got.ID != 1 {
		t.Fatalf("oldest: %d %v", got.ID, err)
	}
	got, err = SelectCompetition(db, 7, 2, now)
	if err != nil || got.ID != 2 {
		t.Fatalf("explicit: %d %v", got.ID, err)
	}
	if _, err = SelectCompetition(db, 8, 2, now); err == nil {
		t.Fatal("explicit selection crossed groups")
	}
	c2.CanPreResult = false
	if err = db.Save(&c2).Error; err != nil {
		t.Fatal(err)
	}
	if _, err = SelectCompetition(db, 7, 2, now); err == nil {
		t.Fatal("explicit selection fell back")
	}
	c.IsDone = true
	if err = db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	if _, err = SelectCompetition(db, 7, 0, now); err == nil {
		t.Fatal("ineligible competition selected")
	}
}

func TestRoundManualAndAdvancement(t *testing.T) {
	db, c, u := fixture(t)
	now := time.Now()
	if err := Refresh(db, c.ID, "333", 2, true, now); err != nil {
		t.Fatal(err)
	}
	ps := []result.PreResults{pending(c, u)}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	if err := Refresh(db, c.ID, "333", 1, false, now); err == nil {
		t.Fatal("closed with pending results")
	}
	if err := Approve(db, c.ID, ps[0].ID, result.DetailOk, u); err != nil {
		t.Fatal(err)
	}
	if err := Refresh(db, c.ID, "333", 1, false, now); err != nil {
		t.Fatal(err)
	}
	var got competition.Competition
	if err := db.First(&got, c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if got.CompJSON.Events[0].Schedule[1].IsRunning {
		t.Fatal("next round automatically opened")
	}
	if err := Refresh(db, c.ID, "333", 2, true, now); err != nil {
		t.Fatal(err)
	}
	if err := Refresh(db, c.ID, "333", 1, true, now); err != nil {
		t.Fatal(err)
	}
	rs := []result.Results{{UserID: 1, Best: 1, EventRoute: event.RouteType1rounds}, {UserID: 2, Best: 2, EventRoute: event.RouteType1rounds}, {UserID: 3, Best: 2, EventRoute: event.RouteType1rounds}, {UserID: 4, Best: 3, EventRoute: event.RouteType1rounds}}
	gotIDs := Advancing(rs, 2)
	slices.Sort(gotIDs)
	if !reflect.DeepEqual(gotIDs, []uint{1, 2, 3}) {
		t.Fatalf("advancement boundary/tie: %v", gotIDs)
	}
}

func TestMultiBlindPendingAndDirectResult(t *testing.T) {
	db, c, u := fixture(t)
	c.CompJSON.Events[0].EventRoute = event.RouteType2RepeatedlyBest
	if err := db.Save(&c).Error; err != nil {
		t.Fatal(err)
	}
	p := pending(c, u)
	p.EventRoute = event.RouteType2RepeatedlyBest
	p.Result = []float64{3, 4, 600}
	ps := []result.PreResults{p}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	if len(ps[0].Result) != 6 || ps[0].Result[3] != 0 || ps[0].Result[4] != 0 || ps[0].Result[5] != result.DNS {
		t.Fatalf("multi blind padding: %v", ps[0].Result)
	}
	if err := Approve(db, c.ID, ps[0].ID, result.DetailOk, u); err != nil {
		t.Fatal(err)
	}
	var saved result.Results
	if err := db.First(&saved, "comp_id = ?", c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.BestRepeatedlyReduction != 3 || saved.BestRepeatedlyTry != 4 || saved.BestRepeatedlyTime != 600 {
		t.Fatalf("wrong multi blind result: %+v", saved)
	}
	// Direct recording uses the same transaction and qualification rules.
	err := db.Transaction(func(tx *gorm.DB) error {
		locked, err := Lock(tx, c.ID)
		if err != nil {
			return err
		}
		_, err = AddResult(tx, locked, u, "333", 1, []float64{4, 4, 500}, nil, time.Now(), true)
		return err
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.First(&saved, "comp_id = ?", c.ID).Error; err != nil {
		t.Fatal(err)
	}
	if saved.BestRepeatedlyReduction != 4 {
		t.Fatal("direct result not updated")
	}
}

func TestApprovalRollbackAndSubmissionChecks(t *testing.T) {
	db, c, u := fixture(t)
	ps := []result.PreResults{pending(c, u)}
	if err := SavePending(db, ps); err != nil {
		t.Fatal(err)
	}
	if err := db.Callback().Update().Before("gorm:update").Register("fail_review", func(tx *gorm.DB) {
		if tx.Statement.Table == "pre_results" {
			tx.AddError(errors.New("injected review failure"))
		}
	}); err != nil {
		t.Fatal(err)
	}
	if err := Approve(db, c.ID, ps[0].ID, result.DetailOk, u); err == nil {
		t.Fatal("review save error ignored")
	}
	var n int64
	if err := db.Model(&result.Results{}).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Fatal("result survived failed approval")
	}
	var p result.PreResults
	if err := db.First(&p, ps[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if p.Finish {
		t.Fatal("pending completed after failed approval")
	}
	for _, tt := range []struct {
		name   string
		change func(*competition.Competition)
	}{
		{"unapproved", func(c *competition.Competition) { c.Status = competition.Reviewing }},
		{"disabled", func(c *competition.Competition) { c.CanPreResult = false }},
	} {
		t.Run(tt.name, func(t *testing.T) {
			changed := c
			changed.CompJSON.Events = append([]competition.CompetitionEvent(nil), c.CompJSON.Events...)
			changed.CompJSON.Events[0].Schedule = append([]competition.Schedule(nil), c.CompJSON.Events[0].Schedule...)
			tt.change(&changed)
			if err := db.Save(&changed).Error; err != nil {
				t.Fatal(err)
			}
			if err := SavePending(db, []result.PreResults{pending(changed, u)}); err == nil {
				t.Fatal("invalid submission accepted")
			}
		})
	}
}

func TestParsingAndGroupMatch(t *testing.T) {
	for _, tt := range []struct {
		name, field, id string
		want            bool
	}{
		{"json", `["123","456"]`, "123", true}, {"substring", "1234,456", "123", false}, {"legacy", "123, 456", "456", true}, {"empty", "", "", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := MatchesGroup(tt.field, tt.id); got != tt.want {
				t.Fatalf("got %v", got)
			}
		})
	}
	if got := SplitBatch("333 1 2 3 4 5 / 333mbf 3/4 10:00\n444 1 2 3 4 5"); len(got) != 3 || got[1] != "333mbf 3/4 10:00" {
		t.Fatalf("batch: %v", got)
	}
	for _, tt := range []struct {
		name    string
		tokens  []string
		route   event.RouteType
		wantErr bool
	}{
		{"times", []string{"1", "2", "d", "s", "5"}, event.RouteType5RoundsAvgHT, false},
		{"multi blind", []string{"3/4", "10:00"}, event.RouteTypeRepeatedly, false},
		{"two multi blind", []string{"3/4", "10:00", "4/4", "9:00"}, event.RouteType2RepeatedlyBest, false},
		{"bad count", []string{"5/4", "10:00"}, event.RouteTypeRepeatedly, true},
		{"missing time", []string{"3/4"}, event.RouteTypeRepeatedly, true},
		{"unknown", []string{"nonsense"}, event.RouteType1rounds, true},
		{"too many", []string{"1", "2"}, event.RouteType1rounds, true},
	} {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseValues(tt.tokens, tt.route)
			if (err != nil) != tt.wantErr {
				t.Fatalf("error: %v", err)
			}
		})
	}
}
