package wca

import (
	"fmt"
	"testing"

	"github.com/guojia99/cubing-pro/src/wca/types"
	jsoniter "github.com/json-iterator/go"
)

func Test_wca_GetPersonResult(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetPersonResult("2018GUOZ01")

	if err != nil {
		t.Fatal(err)
	}

	d, _ := jsoniter.MarshalIndent(out, "", "    ")
	t.Log(string(d))
}

func Test_wca_GetPersonCompetition(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetPersonCompetition("2018GUOZ01")

	if err != nil {
		t.Fatal(err)
	}

	d, _ := jsoniter.MarshalIndent(out, "", "    ")
	t.Log(string(d))
}

func Test_wca_getResultAttemptMap(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	ww := w.(*wca)
	out := ww.getResultAttemptMap([]types.Result{
		{
			ID: 6346720,
		},
	})
	fmt.Println(out)
}

func Test_wca_RankWithEvents(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, _, err := w.GetRankWithEvents(
		[]string{},
		"China",
		true,
		100,
		1,
	)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := jsoniter.MarshalIndent(out, "", "    ")

	fmt.Println(string(d))
}

func Test_wca_GetCountryBestWithEventGroupRank(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetCountryBestWithEventGroupRank("2017XUYO01", true, false)
	if err != nil {
		t.Fatal(err)
	}
	d, _ := jsoniter.MarshalIndent(out, "", "    ")
	fmt.Println(string(d))
}
