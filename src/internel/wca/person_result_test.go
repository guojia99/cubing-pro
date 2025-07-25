package wca

import (
	"fmt"
	"testing"
)

func TestApiGetWCAResults(t *testing.T) {
	t.Run("test", func(t *testing.T) {
		got, err := ApiGetWCAResults("2017XUYO01")

		if err != nil {
			t.Fatal(err)
		}
		fmt.Println(got.PersonName)
		for _, val := range got.Best {
			fmt.Printf("best -> %s %+v\n", val.EventId, val.BestStr)
		}

		for _, val := range got.Avg {
			fmt.Printf("avg -> %s %+v\n", val.EventId, val.AverageStr)
		}
	})

	t.Run("test_cn", func(t *testing.T) {
		got, err := ApiGetWCAResults("徐永浩")

		if err != nil {
			t.Fatal(err)
		}

		fmt.Println(got.PersonName)

		for _, val := range got.Best {
			fmt.Printf("best -> %s %+v\n", val.EventId, val.BestStr)
		}

		for _, val := range got.Avg {
			fmt.Printf("avg -> %s %+v\n", val.EventId, val.AverageStr)
		}
	})
}
