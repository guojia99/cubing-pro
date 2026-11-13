package competition

import (
	"testing"
	"time"
)

func TestIsRunningTimeAt(t *testing.T) {
	parse := func(s string) time.Time {
		v, err := time.Parse(time.RFC3339, s)
		if err != nil {
			t.Fatal(err)
		}
		return v
	}
	c := Competition{Genre: OnlineInformal, GroupID: 1, Status: Running, CompStartTime: parse("2026-11-13T15:00:00+08:00"), CompEndTime: parse("2026-11-14T10:00:00+08:00")}
	for _, tt := range []struct {
		name, now string
		want      bool
	}{
		{"before", "2026-11-12T15:59:59Z", false},
		{"start inclusive", "2026-11-12T16:00:00Z", true},
		{"end day", "2026-11-14T15:59:59Z", true},
		{"next midnight exclusive", "2026-11-14T16:00:00Z", false},
		{"unrelated date", "2027-11-13T00:00:00Z", false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			if got := c.IsRunningTimeAt(parse(tt.now)); got != tt.want {
				t.Fatalf("got %v want %v", got, tt.want)
			}
		})
	}
	c.IsDone = true
	if c.IsRunningTimeAt(parse("2026-11-13T00:00:00+08:00")) {
		t.Fatal("done competition accepted")
	}
	c.IsDone = false
	c.Status = Reviewing
	if c.IsRunningTimeAt(parse("2026-11-13T00:00:00+08:00")) {
		t.Fatal("unapproved competition accepted")
	}
	c.Status = Running
	c.CompStartTime = time.Time{}
	if c.IsRunningTimeAt(time.Now()) {
		t.Fatal("zero dates accepted")
	}
}

func TestNonGroupTimeMatchesHEAD(t *testing.T) {
	now := time.Date(2026, 11, 13, 23, 0, 0, 0, time.FixedZone("other", -7*3600))
	for _, tt := range []struct {
		name      string
		genre     Genre
		group     uint
		wantGroup bool
	}{
		{"group", OnlineInformal, 1, true}, {"ungrouped informal online", OnlineInformal, 0, false}, {"official online", OnlineOfficial, 0, false}, {"official online with group", OnlineOfficial, 1, false}, {"WCA with group", WCA, 1, false}, {"offline", Official, 0, false},
	} {
		t.Run(tt.name, func(t *testing.T) {
			c := Competition{Status: Running, Genre: tt.genre, GroupID: tt.group, IsDone: true}
			if c.IsGroupCompetition() != tt.wantGroup {
				t.Fatal("wrong group boundary")
			}
			if tt.wantGroup {
				if c.IsRunningTimeAt(now) {
					t.Fatal("ended group accepted")
				}
				return
			}
			// HEAD ignores stored start/end dates and IsDone; retain that behavior explicitly.
			start := time.Date(now.Year(), now.Month(), now.Day()-1, 0, 0, 0, 0, now.Location())
			end := time.Date(now.Year(), now.Month(), now.Day()+1, 0, 0, 0, 0, now.Location())
			want := now.After(start) && now.Before(end)
			if c.IsRunningTimeAt(now) != want {
				t.Fatal("non-group timing changed from HEAD")
			}
			c.Status = Reviewing
			if c.IsRunningTimeAt(now) {
				t.Fatal("unapproved accepted")
			}
		})
	}
}
