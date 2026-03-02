package wca

import (
	"encoding/json"
	"testing"
)

func Test_wca_GetEventRankWithTimer(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		false)

	out, count, err := w.GetEventRankWithTimer("333", "China", 2023, true, 1, 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %d", count)

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func Test_wca_GetEventRankWithFullNow(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		false)

	out, count, err := w.GetEventRankWithFullNow(
		"333mbf", "CN", true, 1, 100)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %d", count)

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func Test_wca_GetEventRankWithOnlyYear(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		false)

	out, count, err := w.GetEventRankWithOnlyYear(
		"333", "CN", 2019, false, 1, 20)
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("count: %d", count)

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}

func Test_wca_GetPersonBestRanks(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetPersonBestRanks("2018GUOZ01")
	if err != nil {
		t.Fatal(err)
	}

	dd, _ := json.MarshalIndent(out, "", "\t")
	t.Logf("out: %s", dd)
}
