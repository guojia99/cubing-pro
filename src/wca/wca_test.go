package wca

import (
	"os"
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func Test_wca_GetPersonInfo(t *testing.T) {
	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	//out, err := w.GetPersonInfo("2018GUOZ01")
	//out, err := w.GetPersonInfo("2008DONG06")
	out, err := w.GetPersonInfo("2009ZHEN11")

	if err != nil {
		t.Fatal(err)
	}

	d, _ := jsoniter.MarshalIndent(out, "", "    ")
	t.Log(string(d))

	os.WriteFile("test.json", d, 0644)
}
