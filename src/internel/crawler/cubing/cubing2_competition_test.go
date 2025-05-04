package cubing

import (
	"fmt"
	"testing"
)

func TestDCubingCompetition_GetNewCompetitions(t *testing.T) {
	c := NewDCubingCompetition()
	fmt.Println(c.GetNewCompetitions())
}

func TestDCubingCompetition_getPage(t *testing.T) {
	c := NewDCubingCompetition()
	out, find, err := c.getPage("Zhuhai-Open-2025", "https://cubing.com/competition/Zhuhai-Open-2025")
	//out, find, err := c.getPage("Yancheng-Open-2025", "https://cubing.com/competition/Yancheng-Open-2025")

	if err != nil {
		t.Fatal(err)
	}
	t.Logf("%v", find)
	t.Logf("%+v", out)
}
