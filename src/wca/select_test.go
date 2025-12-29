package wca

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func Test_wca_GetPersonResult(t *testing.T) {
	w, _ := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetPersonResult("2018GUOZ01")

	if err != nil {
		t.Fatal(err)
	}

	d, _ := jsoniter.MarshalIndent(out, "", "    ")
	t.Log(string(d))
}

func Test_wca_GetPersonCompetition(t *testing.T) {
	w, _ := NewWCA(
		"root@tcp(127.0.0.1:33306)/",
		"/home/guojia/cubingPro/wca_db",
		"/home/guojia/cubingPro/wca_db/sync_path",
		false)

	out, err := w.GetPersonCompetition("2018GUOZ01")

	if err != nil {
		t.Fatal(err)
	}

	d, _ := jsoniter.MarshalIndent(out, "", "    ")
	t.Log(string(d))
}
