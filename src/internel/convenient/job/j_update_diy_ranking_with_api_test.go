package job

import (
	"fmt"
	"testing"
)

func TestUpdateDiyRankings_apiGetWCAResults(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		u := &UpdateDiyRankings{}
		got, err := u.apiGetWCAResults("2017XUYO01")

		if err != nil {
			t.Fatal(err)
		}

		for _, val := range got.Best {
			fmt.Printf("best -> %s %+v\n", val.EventId, val.BestStr)
		}

		for _, val := range got.Avg {
			fmt.Printf("avg -> %s %+v\n", val.EventId, val.AverageStr)
		}
	})

}

func TestUpdateDiyRankings_apiGetAllResult(t *testing.T) {
	u := &UpdateDiyRankings{}

	id := []string{
		"2018GUOZ01",
		"2018XUEZ01",
		"2019LIUY06",
		"2017XUZI03",
		"2023GUXI01",
		"2021HUAN08",
		"2023ZHEN26",
		"2017XUYO01",
		"2017LIUG02",
		"2024ZHAN08",
		"2024LUOW02",
		"2017XUYO01",
	}

	out := u.apiGetAllResult(id)
	fmt.Println(out)

}
