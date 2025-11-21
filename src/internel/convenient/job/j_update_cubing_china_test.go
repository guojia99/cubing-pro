package job

import (
	"testing"

	jsoniter "github.com/json-iterator/go"
)

func Test_getCompetitionWithPage(t *testing.T) {
	out, count, err := getCompetitionWithPage(1)
	if err != nil {
		t.Fatal(err)
	}

	data, _ := jsoniter.MarshalIndent(out, "", "    ")
	t.Log(string(data))
	t.Logf("------------ %d", count)
}

func TestUpdateCubingChinaComps_Run(t *testing.T) {
	u := UpdateCubingChinaComps{}

	if err := u.Run(); err != nil {
		t.Fatal(err)
	}
}
