package wca

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/guojia99/cubing-pro/src/wca/types"
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

func Test_FMC_Length(t *testing.T) {

	w := NewWCA(
		"root@tcp(127.0.0.1:33036)/",
		"/Users/guojia/data/cubingPro/wca_db",
		"/Users/guojia/data/cubingPro/wca_db/sync_path",
		false)

	ww := w.(*wca)

	var sc []types.Scramble
	ww.db.Where("event_id = ?", "333fm").Find(&sc)
	var out []string
	var numMap = make(map[int]int)

	var lastFour = make(map[string]int)

	for _, s := range sc {
		if strings.HasPrefix(s.Scramble, "R' U' F") {
			out = append(out, s.Scramble)

			stp := strings.Split(s.Scramble, " ")
			lent := len(stp)
			if _, ok := numMap[lent]; !ok {
				numMap[lent] = 0
			}
			numMap[lent] += 1

			last := strings.Join(stp[lent-4:], " ")
			if _, ok := lastFour[last]; !ok {
				lastFour[last] = 0
			}
			lastFour[last] += 1
		}
	}

	d, _ := jsoniter.MarshalIndent(numMap, "", "    ")
	fmt.Println(string(d))

	d, _ = jsoniter.MarshalIndent(lastFour, "", "    ")
	fmt.Println(string(d))

}
